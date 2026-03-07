package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Agent is the Schema for the agents API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Role",type=string,JSONPath=`.spec.role`
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.llm.provider`
// +kubebuilder:printcolumn:name="Model",type=string,JSONPath=`.spec.llm.model`
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Agent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentSpec   `json:"spec,omitempty"`
	Status AgentStatus `json:"status,omitempty"`
}

// AgentSpec defines the desired state of Agent
// +kubebuilder:object:generate=true
type AgentSpec struct {
	Role        string                      `json:"role"`
	Goal        string                      `json:"goal,omitempty"`
	LLM         LLMSpec                     `json:"llm"`
	Tools       []string                    `json:"tools,omitempty"`
	Memory      *MemorySpec                 `json:"memory,omitempty"`
	Replicas    int                         `json:"replicas,omitempty"`
	Resources   corev1.ResourceRequirements `json:"resources,omitempty"`
	Autoscaling *AutoscalingSpec            `json:"autoscaling,omitempty"`
}

// LLMSpec defines LLM configuration
// +kubebuilder:object:generate=true
type LLMSpec struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"maxTokens,omitempty"`
}

// MemorySpec defines memory configuration
// +kubebuilder:object:generate=true
type MemorySpec struct {
	Enabled  bool              `json:"enabled"`
	Provider string            `json:"provider,omitempty"`
	Config   map[string]string `json:"config,omitempty"`
}

// AutoscalingSpec defines autoscaling configuration
// +kubebuilder:object:generate=true
type AutoscalingSpec struct {
	Enabled              bool `json:"enabled"`
	MinReplicas          int  `json:"minReplicas,omitempty"`
	MaxReplicas          int  `json:"maxReplicas,omitempty"`
	TargetCPUUtilization int  `json:"targetCPUUtilization,omitempty"`
}

// AgentStatus defines the observed state of Agent
// +kubebuilder:object:generate=true
type AgentStatus struct {
	Phase         string             `json:"phase,omitempty"`
	Replicas      int32              `json:"replicas,omitempty"`
	ReadyReplicas int32              `json:"readyReplicas,omitempty"`
	Conditions    []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// AgentList contains a list of Agent
type AgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Agent `json:"items"`
}

// Workflow is the Schema for the workflows API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec,omitempty"`
	Status WorkflowStatus `json:"status,omitempty"`
}

// WorkflowSpec defines the desired state of Workflow
// +kubebuilder:object:generate=true
type WorkflowSpec struct {
	Name          string         `json:"name,omitempty"`
	Description   string         `json:"description,omitempty"`
	Steps         []WorkflowStep `json:"steps"`
	Schedule      string         `json:"schedule,omitempty"`
	Timeout       string         `json:"timeout,omitempty"`
	OnFailure     string         `json:"onFailure,omitempty"`
	MaxConcurrent int            `json:"maxConcurrent,omitempty"`
}

// WorkflowStep defines a step in the workflow
// +kubebuilder:object:generate=true
type WorkflowStep struct {
	Name            string   `json:"name"`
	Type            string   `json:"type,omitempty"` // "agent" (default) or "approval"
	Agent           string   `json:"agent"`
	Goal            string   `json:"goal"`
	OutputKey       string   `json:"outputKey,omitempty"`
	Condition       string   `json:"condition,omitempty"`
	DependsOn       []string `json:"dependsOn,omitempty"`
	Timeout         string   `json:"timeout,omitempty"`
	Retry           int      `json:"retry,omitempty"`
	Approvers       []string `json:"approvers,omitempty"`       // for type=approval
	ApprovalMessage string   `json:"approvalMessage,omitempty"` // for type=approval
}

// WorkflowStatus defines the observed state of Workflow
// +kubebuilder:object:generate=true
type WorkflowStatus struct {
	Phase            string             `json:"phase,omitempty"`
	Message          string             `json:"message,omitempty"`
	StartTime        *metav1.Time       `json:"startTime,omitempty"`
	CompletionTime   *metav1.Time       `json:"completionTime,omitempty"`
	CurrentStep      string             `json:"currentStep,omitempty"`
	CompletedSteps   []string           `json:"completedSteps,omitempty"`
	FailedSteps      []string           `json:"failedSteps,omitempty"`
	LastScheduleTime *metav1.Time       `json:"lastScheduleTime,omitempty"`
	NextScheduleTime *metav1.Time       `json:"nextScheduleTime,omitempty"`
	Conditions       []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// WorkflowList contains a list of Workflow
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
}

// Policy is the Schema for the policies API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicySpec   `json:"spec,omitempty"`
	Status PolicyStatus `json:"status,omitempty"`
}

// PolicySpec defines the desired state of Policy
// +kubebuilder:object:generate=true
type PolicySpec struct {
	Description string       `json:"description,omitempty"`
	Version     string       `json:"version,omitempty"`
	Rules       []PolicyRule `json:"rules"`
	Approvers   []string     `json:"approvers,omitempty"`
	Enabled     bool         `json:"enabled"`
}

// PolicyRule defines a policy rule
// +kubebuilder:object:generate=true
type PolicyRule struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	Condition   string `json:"condition"`
	Action      string `json:"action"`
	Severity    string `json:"severity,omitempty"`
}

// PolicyStatus defines the observed state of Policy
// +kubebuilder:object:generate=true
type PolicyStatus struct {
	Active      bool               `json:"active,omitempty"`
	LastApplied *metav1.Time       `json:"lastApplied,omitempty"`
	Violations  int                `json:"violations,omitempty"`
	Approvals   int                `json:"approvals,omitempty"`
	Conditions  []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// PolicyList contains a list of Policy
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}

// Note: DeepCopy methods are generated by controller-gen. Run: make generate-operator

func init() {
	SchemeBuilder.Register(&Agent{}, &AgentList{})
	SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
	SchemeBuilder.Register(&Policy{}, &PolicyList{})
}
