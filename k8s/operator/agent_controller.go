package operator

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	agentruntimev1 "github.com/NikoSokratous/unagnt/k8s/operator/api/v1"
)

// AgentReconciler reconciles an Agent object
type AgentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=agentruntime.io,resources=agents,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agentruntime.io,resources=agents/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=agentruntime.io,resources=agents/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *AgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Agent instance
	agent := &agentruntimev1.Agent{}
	if err := r.Get(ctx, req.NamespacedName, agent); err != nil {
		logger.Error(err, "unable to fetch Agent")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create or update Deployment
	deployment := r.deploymentForAgent(agent)
	if err := r.createOrUpdateDeployment(ctx, agent, deployment); err != nil {
		logger.Error(err, "failed to create or update Deployment")
		return ctrl.Result{}, err
	}

	// Create or update Service
	service := r.serviceForAgent(agent)
	if err := r.createOrUpdateService(ctx, agent, service); err != nil {
		logger.Error(err, "failed to create or update Service")
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.updateAgentStatus(ctx, agent); err != nil {
		logger.Error(err, "failed to update Agent status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// deploymentForAgent returns a Deployment object for the Agent
func (r *AgentReconciler) deploymentForAgent(agent *agentruntimev1.Agent) *appsv1.Deployment {
	labels := labelsForAgent(agent)
	replicas := int32(agent.Spec.Replicas)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.Name,
			Namespace: agent.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "agent",
							Image: "agentruntime/agent:latest",
							Env: []corev1.EnvVar{
								{
									Name:  "AGENT_ROLE",
									Value: agent.Spec.Role,
								},
								{
									Name:  "AGENT_GOAL",
									Value: agent.Spec.Goal,
								},
								{
									Name:  "LLM_PROVIDER",
									Value: agent.Spec.LLM.Provider,
								},
								{
									Name:  "LLM_MODEL",
									Value: agent.Spec.LLM.Model,
								},
								{
									Name:  "LLM_TEMPERATURE",
									Value: fmt.Sprintf("%f", agent.Spec.LLM.Temperature),
								},
								{
									Name:  "LLM_MAX_TOKENS",
									Value: fmt.Sprintf("%d", agent.Spec.LLM.MaxTokens),
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									Name:          "http",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: agent.Spec.Resources.Requests,
								Limits:   agent.Spec.Resources.Limits,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       20,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
						},
					},
				},
			},
		},
	}

	// Set Agent as the owner
	ctrl.SetControllerReference(agent, deployment, r.Scheme)
	return deployment
}

// serviceForAgent returns a Service object for the Agent
func (r *AgentReconciler) serviceForAgent(agent *agentruntimev1.Agent) *corev1.Service {
	labels := labelsForAgent(agent)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.Name,
			Namespace: agent.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
					Name:       "http",
				},
			},
		},
	}

	ctrl.SetControllerReference(agent, service, r.Scheme)
	return service
}

// createOrUpdateDeployment creates or updates a Deployment
func (r *AgentReconciler) createOrUpdateDeployment(ctx context.Context, agent *agentruntimev1.Agent, deployment *appsv1.Deployment) error {
	found := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKeyFromObject(deployment), found)

	if err != nil && client.IgnoreNotFound(err) == nil {
		// Create new deployment
		if err := r.Create(ctx, deployment); err != nil {
			return fmt.Errorf("failed to create Deployment: %w", err)
		}
	} else if err == nil {
		// Update existing deployment
		found.Spec = deployment.Spec
		if err := r.Update(ctx, found); err != nil {
			return fmt.Errorf("failed to update Deployment: %w", err)
		}
	} else {
		return err
	}

	return nil
}

// createOrUpdateService creates or updates a Service
func (r *AgentReconciler) createOrUpdateService(ctx context.Context, agent *agentruntimev1.Agent, service *corev1.Service) error {
	found := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKeyFromObject(service), found)

	if err != nil && client.IgnoreNotFound(err) == nil {
		// Create new service
		if err := r.Create(ctx, service); err != nil {
			return fmt.Errorf("failed to create Service: %w", err)
		}
	} else if err == nil {
		// Update existing service
		found.Spec.Ports = service.Spec.Ports
		found.Spec.Selector = service.Spec.Selector
		if err := r.Update(ctx, found); err != nil {
			return fmt.Errorf("failed to update Service: %w", err)
		}
	} else {
		return err
	}

	return nil
}

// updateAgentStatus updates the Agent status
func (r *AgentReconciler) updateAgentStatus(ctx context.Context, agent *agentruntimev1.Agent) error {
	// Get the Deployment
	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{Name: agent.Name, Namespace: agent.Namespace}, deployment); err != nil {
		return err
	}

	// Update status based on deployment
	agent.Status.Replicas = deployment.Status.Replicas
	agent.Status.ReadyReplicas = deployment.Status.ReadyReplicas

	if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
		agent.Status.Phase = "Running"
	} else if deployment.Status.ReadyReplicas > 0 {
		agent.Status.Phase = "Pending"
	} else {
		agent.Status.Phase = "Failed"
	}

	// Update the status
	if err := r.Status().Update(ctx, agent); err != nil {
		return fmt.Errorf("failed to update Agent status: %w", err)
	}

	return nil
}

// labelsForAgent returns labels for the Agent
func labelsForAgent(agent *agentruntimev1.Agent) map[string]string {
	return map[string]string{
		"app":                      "agentruntime",
		"agentruntime.io/agent":    agent.Name,
		"agentruntime.io/role":     agent.Spec.Role,
		"agentruntime.io/provider": agent.Spec.LLM.Provider,
	}
}

// SetupWithManager sets up the controller with the Manager
func (r *AgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentruntimev1.Agent{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
