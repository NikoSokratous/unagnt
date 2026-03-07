package operator

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	agentruntimev1 "github.com/NikoSokratous/unagnt/k8s/operator/api/v1"
)

// WorkflowReconciler reconciles a Workflow object
type WorkflowReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=unagnt.io,resources=workflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=unagnt.io,resources=workflows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=unagnt.io,resources=workflows/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *WorkflowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Workflow instance
	workflow := &agentruntimev1.Workflow{}
	if err := r.Get(ctx, req.NamespacedName, workflow); err != nil {
		logger.Error(err, "unable to fetch Workflow")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle scheduled workflows
	if workflow.Spec.Schedule != "" {
		return r.handleScheduledWorkflow(ctx, workflow)
	}

	// Handle one-time workflow execution
	return r.handleWorkflowExecution(ctx, workflow)
}

// handleScheduledWorkflow handles scheduled workflow execution
func (r *WorkflowReconciler) handleScheduledWorkflow(ctx context.Context, workflow *agentruntimev1.Workflow) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if workflow.Spec.Schedule == "" {
		return ctrl.Result{}, nil
	}

	// Parse cron schedule
	schedule, err := cron.ParseStandard(workflow.Spec.Schedule)
	if err != nil {
		logger.Error(err, "Invalid cron schedule", "schedule", workflow.Spec.Schedule)
		workflow.Status.Phase = "Failed"
		workflow.Status.Message = fmt.Sprintf("Invalid cron schedule: %v", err)
		if updateErr := r.Status().Update(ctx, workflow); updateErr != nil {
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{}, err
	}

	logger.Info("Scheduled workflow", "schedule", workflow.Spec.Schedule)

	now := time.Now()

	// Calculate next run time
	nextTime := schedule.Next(now)
	nextScheduleTime := metav1.NewTime(nextTime)
	workflow.Status.NextScheduleTime = &nextScheduleTime

	// Check if it's time to run
	shouldRun := false
	if workflow.Status.LastScheduleTime == nil {
		// First run
		shouldRun = true
	} else {
		lastRun := workflow.Status.LastScheduleTime.Time
		expectedNextRun := schedule.Next(lastRun)

		// If current time is past the expected next run, execute now
		if now.After(expectedNextRun) || now.Equal(expectedNextRun) {
			shouldRun = true
		}
	}

	if shouldRun {
		// Update last schedule time
		lastScheduleTime := metav1.Now()
		workflow.Status.LastScheduleTime = &lastScheduleTime

		// Execute the workflow
		if _, err := r.handleWorkflowExecution(ctx, workflow); err != nil {
			logger.Error(err, "Failed to execute scheduled workflow")
			// Don't return error - continue scheduling
		}
	}

	// Update status
	if err := r.Status().Update(ctx, workflow); err != nil {
		return ctrl.Result{}, err
	}

	// Calculate requeue delay (check every minute or before next run)
	requeueAfter := time.Until(nextTime)
	if requeueAfter > time.Minute {
		requeueAfter = time.Minute
	}
	if requeueAfter < 0 {
		requeueAfter = 10 * time.Second
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

// handleWorkflowExecution handles workflow execution
func (r *WorkflowReconciler) handleWorkflowExecution(ctx context.Context, workflow *agentruntimev1.Workflow) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Check if workflow is already running
	if workflow.Status.Phase == "Running" {
		return r.monitorWorkflowExecution(ctx, workflow)
	}

	// Start new workflow execution
	if workflow.Status.Phase == "" || workflow.Status.Phase == "Pending" {
		return r.startWorkflowExecution(ctx, workflow)
	}

	// Workflow is complete
	logger.Info("Workflow complete", "phase", workflow.Status.Phase)
	return ctrl.Result{}, nil
}

// startWorkflowExecution starts a new workflow execution
func (r *WorkflowReconciler) startWorkflowExecution(ctx context.Context, workflow *agentruntimev1.Workflow) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Create ConfigMap with workflow definition
	configMap := r.configMapForWorkflow(workflow)
	if err := r.createOrUpdateConfigMap(ctx, workflow, configMap); err != nil {
		logger.Error(err, "failed to create ConfigMap")
		return ctrl.Result{}, err
	}

	// Create Job to execute workflow
	job := r.jobForWorkflow(workflow)
	if err := r.Create(ctx, job); err != nil {
		logger.Error(err, "failed to create Job")
		return ctrl.Result{}, err
	}

	// Update status
	now := metav1.Now()
	workflow.Status.Phase = "Running"
	workflow.Status.StartTime = &now
	workflow.Status.CurrentStep = workflow.Spec.Steps[0].Name
	workflow.Status.CompletedSteps = []string{}
	workflow.Status.FailedSteps = []string{}

	if err := r.Status().Update(ctx, workflow); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Started workflow execution", "job", job.Name)
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

// monitorWorkflowExecution monitors an active workflow execution
func (r *WorkflowReconciler) monitorWorkflowExecution(ctx context.Context, workflow *agentruntimev1.Workflow) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Get the Job
	job := &batchv1.Job{}
	if err := r.Get(ctx, client.ObjectKey{Name: workflow.Name, Namespace: workflow.Namespace}, job); err != nil {
		logger.Error(err, "failed to get Job")
		return ctrl.Result{}, err
	}

	// Check job status
	if job.Status.Succeeded > 0 {
		now := metav1.Now()
		workflow.Status.Phase = "Succeeded"
		workflow.Status.CompletionTime = &now
		workflow.Status.CurrentStep = ""

		if err := r.Status().Update(ctx, workflow); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info("Workflow succeeded")
		return ctrl.Result{}, nil
	}

	if job.Status.Failed > 0 {
		now := metav1.Now()
		workflow.Status.Phase = "Failed"
		workflow.Status.CompletionTime = &now

		if err := r.Status().Update(ctx, workflow); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info("Workflow failed")
		return ctrl.Result{}, nil
	}

	// Still running
	logger.Info("Workflow still running", "active", job.Status.Active)
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

// configMapForWorkflow creates a ConfigMap with workflow definition
func (r *WorkflowReconciler) configMapForWorkflow(workflow *agentruntimev1.Workflow) *corev1.ConfigMap {
	// Convert workflow to YAML-like format
	workflowData := fmt.Sprintf(`name: %s
description: %s
steps:
`, workflow.Spec.Name, workflow.Spec.Description)

	for _, step := range workflow.Spec.Steps {
		stepType := step.Type
		if stepType == "" {
			stepType = "agent"
		}
		agent, goal := step.Agent, step.Goal
		if stepType == "approval" && agent == "" {
			agent, goal = "approval", "human sign-off"
		}
		workflowData += fmt.Sprintf(`  - name: %s
    type: %s
    agent: %s
    goal: "%s"
`, step.Name, stepType, agent, goal)

		if step.OutputKey != "" {
			workflowData += fmt.Sprintf("    output_key: %s\n", step.OutputKey)
		}
		if step.Condition != "" {
			workflowData += fmt.Sprintf("    condition: \"%s\"\n", step.Condition)
		}
		if len(step.DependsOn) > 0 {
			workflowData += "    depends_on:\n"
			for _, dep := range step.DependsOn {
				workflowData += fmt.Sprintf("      - %s\n", dep)
			}
		}
		if step.Type == "approval" {
			if len(step.Approvers) > 0 {
				workflowData += "    approvers:\n"
				for _, a := range step.Approvers {
					workflowData += fmt.Sprintf("      - %s\n", a)
				}
			}
			if step.ApprovalMessage != "" {
				workflowData += fmt.Sprintf("    approval_message: \"%s\"\n", step.ApprovalMessage)
			}
		}
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-workflow", workflow.Name),
			Namespace: workflow.Namespace,
			Labels: map[string]string{
				"app":                "agentruntime",
				"unagnt.io/workflow": workflow.Name,
			},
		},
		Data: map[string]string{
			"workflow.yaml": workflowData,
		},
	}

	ctrl.SetControllerReference(workflow, configMap, r.Scheme)
	return configMap
}

// jobForWorkflow creates a Job to execute the workflow
func (r *WorkflowReconciler) jobForWorkflow(workflow *agentruntimev1.Workflow) *batchv1.Job {
	labels := map[string]string{
		"app":                "unagnt",
		"unagnt.io/workflow": workflow.Name,
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workflow.Name,
			Namespace: workflow.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:  "workflow-executor",
							Image: "unagnt/executor:latest",
							Args: []string{
								"workflow",
								"run",
								"/config/workflow.yaml",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workflow-config",
									MountPath: "/config",
									ReadOnly:  true,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "WORKFLOW_NAME",
									Value: workflow.Spec.Name,
								},
								{
									Name:  "WORKFLOW_NAMESPACE",
									Value: workflow.Namespace,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "workflow-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: fmt.Sprintf("%s-workflow", workflow.Name),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	ctrl.SetControllerReference(workflow, job, r.Scheme)
	return job
}

// createOrUpdateConfigMap creates or updates a ConfigMap
func (r *WorkflowReconciler) createOrUpdateConfigMap(ctx context.Context, workflow *agentruntimev1.Workflow, configMap *corev1.ConfigMap) error {
	found := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKeyFromObject(configMap), found)

	if err != nil && client.IgnoreNotFound(err) == nil {
		if err := r.Create(ctx, configMap); err != nil {
			return fmt.Errorf("failed to create ConfigMap: %w", err)
		}
	} else if err == nil {
		found.Data = configMap.Data
		if err := r.Update(ctx, found); err != nil {
			return fmt.Errorf("failed to update ConfigMap: %w", err)
		}
	} else {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *WorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentruntimev1.Workflow{}).
		Owns(&batchv1.Job{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
