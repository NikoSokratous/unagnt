package operator

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	agentruntimev1 "github.com/NikoSokratous/unagnt/k8s/operator/api/v1"
)

// PolicyReconciler reconciles a Policy object
type PolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=agentruntime.io,resources=policies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agentruntime.io,resources=policies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=agentruntime.io,resources=policies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop
func (r *PolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Policy instance
	policy := &agentruntimev1.Policy{}
	if err := r.Get(ctx, req.NamespacedName, policy); err != nil {
		logger.Error(err, "unable to fetch Policy")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Validate policy rules
	if err := r.validatePolicy(policy); err != nil {
		logger.Error(err, "policy validation failed")
		policy.Status.Active = false
		if err := r.Status().Update(ctx, policy); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	// Apply policy if enabled
	if policy.Spec.Enabled {
		if err := r.applyPolicy(ctx, policy); err != nil {
			logger.Error(err, "failed to apply policy")
			return ctrl.Result{}, err
		}
	} else {
		policy.Status.Active = false
	}

	// Update status
	if err := r.updatePolicyStatus(ctx, policy); err != nil {
		logger.Error(err, "failed to update policy status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// validatePolicy validates the policy rules
func (r *PolicyReconciler) validatePolicy(policy *agentruntimev1.Policy) error {
	if len(policy.Spec.Rules) == 0 {
		return fmt.Errorf("policy must have at least one rule")
	}

	// Validate each rule
	for _, rule := range policy.Spec.Rules {
		if rule.ID == "" {
			return fmt.Errorf("rule must have an ID")
		}
		if rule.Condition == "" {
			return fmt.Errorf("rule %s must have a condition", rule.ID)
		}
		if rule.Action == "" {
			return fmt.Errorf("rule %s must have an action", rule.ID)
		}

		// Validate action type
		validActions := map[string]bool{
			"allow":            true,
			"deny":             true,
			"warn":             true,
			"require_approval": true,
		}
		if !validActions[rule.Action] {
			return fmt.Errorf("rule %s has invalid action: %s", rule.ID, rule.Action)
		}
	}

	return nil
}

// applyPolicy applies the policy to the cluster
func (r *PolicyReconciler) applyPolicy(ctx context.Context, policy *agentruntimev1.Policy) error {
	logger := log.FromContext(ctx)

	// In a real implementation, this would:
	// 1. Register policy rules with the policy engine
	// 2. Set up webhooks for policy enforcement
	// 3. Configure audit logging

	logger.Info("Applied policy", "name", policy.Name, "rules", len(policy.Spec.Rules))

	return nil
}

// updatePolicyStatus updates the policy status
func (r *PolicyReconciler) updatePolicyStatus(ctx context.Context, policy *agentruntimev1.Policy) error {
	now := metav1.Now()
	policy.Status.Active = policy.Spec.Enabled
	policy.Status.LastApplied = &now

	if err := r.Status().Update(ctx, policy); err != nil {
		return fmt.Errorf("failed to update policy status: %w", err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *PolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentruntimev1.Policy{}).
		Complete(r)
}
