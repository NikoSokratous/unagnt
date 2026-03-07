package orchestrate

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/workflow"
	"gopkg.in/yaml.v3"
)

// WorkflowConfig defines a multi-agent workflow.
type WorkflowConfig struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Steps       []WorkflowStep `yaml:"steps,omitempty"`
	Parallel    []WorkflowStep `yaml:"parallel,omitempty"`
	Aggregate   bool           `yaml:"aggregate"`
	Timeout     string         `yaml:"timeout"`
	OnError     string         `yaml:"on_error"` // "stop" or "continue"
}

// WorkflowStep defines a single step in a workflow.
type WorkflowStep struct {
	Name      string   `yaml:"name"`
	Agent     string   `yaml:"agent"`
	Goal      string   `yaml:"goal"`
	OutputKey string   `yaml:"output_key,omitempty"`
	Condition string   `yaml:"condition,omitempty"`  // CEL expression
	DependsOn []string `yaml:"depends_on,omitempty"` // For DAG workflows
	Timeout   string   `yaml:"timeout,omitempty"`
	Retry     int      `yaml:"retry,omitempty"`
}

// WorkflowResult contains the result of a workflow execution.
type WorkflowResult struct {
	WorkflowName string                 `json:"workflow_name"`
	Status       string                 `json:"status"` // completed, failed, partial
	Steps        []StepResult           `json:"steps"`
	Outputs      map[string]interface{} `json:"outputs"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  time.Time              `json:"completed_at"`
	Duration     time.Duration          `json:"duration"`
	Error        string                 `json:"error,omitempty"`
}

// StepResult contains the result of a single workflow step.
type StepResult struct {
	Name        string        `json:"name"`
	Agent       string        `json:"agent"`
	Status      string        `json:"status"` // completed, failed, skipped
	RunID       string        `json:"run_id,omitempty"`
	Output      interface{}   `json:"output,omitempty"`
	Error       string        `json:"error,omitempty"`
	Duration    time.Duration `json:"duration"`
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt time.Time     `json:"completed_at"`
}

// WorkflowEngine executes multi-agent workflows.
type WorkflowEngine struct {
	server       *Server
	celEvaluator *workflow.CELEvaluator
	executor     StepExecutor
}

// NewWorkflowEngine creates a new workflow engine.
func NewWorkflowEngine(server *Server) *WorkflowEngine {
	celEval, _ := workflow.NewCELEvaluator()
	return &WorkflowEngine{
		server:       server,
		celEvaluator: celEval,
		executor: &RuntimeStepExecutor{
			AllowSimulatedFallback: false,
			StorePath:              "agent.db",
		},
	}
}

// NewWorkflowEngineWithExecutor creates a workflow engine with a custom step executor.
func NewWorkflowEngineWithExecutor(server *Server, executor StepExecutor) *WorkflowEngine {
	celEval, _ := workflow.NewCELEvaluator()
	eng := &WorkflowEngine{
		server:       server,
		celEvaluator: celEval,
		executor:     executor,
	}
	if eng.executor == nil {
		eng.executor = SimulatedExecutor{}
	}
	return eng
}

// Execute runs a workflow.
func (w *WorkflowEngine) Execute(ctx context.Context, config *WorkflowConfig) (*WorkflowResult, error) {
	result := &WorkflowResult{
		WorkflowName: config.Name,
		Status:       "running",
		Outputs:      make(map[string]interface{}),
		StartedAt:    time.Now(),
	}

	// Parse timeout
	timeout := 300 * time.Second
	if config.Timeout != "" {
		if d, err := time.ParseDuration(config.Timeout); err == nil {
			timeout = d
		}
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute based on workflow type
	var err error
	if len(config.Steps) > 0 {
		// Sequential workflow
		err = w.executeSequential(execCtx, config, result)
	} else if len(config.Parallel) > 0 {
		// Parallel workflow
		err = w.executeParallel(execCtx, config, result)
	} else {
		err = fmt.Errorf("workflow must have either steps or parallel configuration")
	}

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		return result, err
	}

	// Check if all steps completed
	allCompleted := true
	for _, step := range result.Steps {
		if step.Status != "completed" {
			allCompleted = false
			break
		}
	}

	if allCompleted {
		result.Status = "completed"
	} else {
		result.Status = "partial"
	}

	return result, nil
}

// executeSequential runs workflow steps in sequence.
func (w *WorkflowEngine) executeSequential(ctx context.Context, config *WorkflowConfig, result *WorkflowResult) error {
	for i, step := range config.Steps {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check condition if specified
		if step.Condition != "" {
			shouldExecute, err := w.evaluateCondition(step.Condition, result.Outputs)
			if err != nil {
				return fmt.Errorf("evaluate condition for step %s: %w", step.Name, err)
			}
			if !shouldExecute {
				// Skip this step
				result.Steps = append(result.Steps, StepResult{
					Name:        step.Name,
					Status:      "skipped",
					StartedAt:   time.Now(),
					CompletedAt: time.Now(),
				})
				continue
			}
		}

		// Execute step
		stepResult, err := w.executeStep(ctx, &step, result.Outputs)
		result.Steps = append(result.Steps, *stepResult)

		if err != nil {
			// Handle error based on on_error policy
			if config.OnError == "continue" {
				// Continue to next step
				continue
			}
			// Stop on error (default)
			return fmt.Errorf("step %d (%s) failed: %w", i, step.Name, err)
		}

		// Store output if output_key is specified
		if step.OutputKey != "" && stepResult.Output != nil {
			result.Outputs[step.OutputKey] = stepResult.Output
		}
	}

	return nil
}

// executeParallel runs workflow steps in parallel.
func (w *WorkflowEngine) executeParallel(ctx context.Context, config *WorkflowConfig, result *WorkflowResult) error {
	var wg sync.WaitGroup
	var mu sync.Mutex

	stepResults := make([]*StepResult, len(config.Parallel))
	errors := make([]error, len(config.Parallel))

	// Launch all steps in parallel
	for i := range config.Parallel {
		wg.Add(1)

		go func(idx int, step WorkflowStep) {
			defer wg.Done()

			// Execute step
			stepResult, err := w.executeStep(ctx, &step, result.Outputs)

			mu.Lock()
			stepResults[idx] = stepResult
			errors[idx] = err
			mu.Unlock()
		}(i, config.Parallel[i])
	}

	// Wait for all to complete
	wg.Wait()

	// Collect results
	mu.Lock()
	result.Steps = make([]StepResult, 0, len(stepResults))
	for i, sr := range stepResults {
		if sr != nil {
			result.Steps = append(result.Steps, *sr)

			// Store output if output_key is specified
			if config.Parallel[i].OutputKey != "" && sr.Output != nil {
				result.Outputs[config.Parallel[i].OutputKey] = sr.Output
			}
		}
	}
	mu.Unlock()

	// Check for errors
	if config.OnError != "continue" {
		for i, err := range errors {
			if err != nil {
				return fmt.Errorf("parallel step %s failed: %w", config.Parallel[i].Name, err)
			}
		}
	}

	return nil
}

// executeStep executes a single workflow step via the configured executor.
func (w *WorkflowEngine) executeStep(ctx context.Context, step *WorkflowStep, outputs map[string]interface{}) (*StepResult, error) {
	goal := w.renderGoalWithOutputs(step.Goal, outputs)
	exec := w.executor
	if exec == nil {
		exec = SimulatedExecutor{}
	}
	stepResult, err := exec.ExecuteStep(ctx, step.Agent, goal, outputs)
	if err != nil {
		now := time.Now()
		return &StepResult{
			Name:        step.Name,
			Agent:       step.Agent,
			Status:      "failed",
			Error:       err.Error(),
			StartedAt:   now,
			CompletedAt: now,
		}, err
	}
	stepResult.Name = step.Name
	return stepResult, nil
}

// renderGoalWithOutputs renders a goal template with workflow outputs.
func (w *WorkflowEngine) renderGoalWithOutputs(goalTemplate string, outputs map[string]interface{}) string {
	// Use Go's text/template for proper template rendering
	tmpl, err := template.New("goal").Parse(goalTemplate)
	if err != nil {
		// If parsing fails, return the original template
		return goalTemplate
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, map[string]interface{}{
		"Outputs": outputs,
		"outputs": outputs, // lowercase alias
	}); err != nil {
		// If execution fails, return the original template
		return goalTemplate
	}

	return result.String()
}

// evaluateCondition evaluates a CEL condition with workflow context
func (w *WorkflowEngine) evaluateCondition(condition string, outputs map[string]interface{}) (bool, error) {
	if w.celEvaluator == nil {
		return true, nil // If no evaluator, always execute
	}

	// Build context for CEL evaluation
	context := map[string]interface{}{
		"outputs": outputs,
	}

	return w.celEvaluator.Evaluate(condition, context)
}

// LoadWorkflow loads a workflow configuration from YAML file.
func LoadWorkflow(path string) (*WorkflowConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow file: %w", err)
	}

	var workflow WorkflowConfig
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("parse workflow: %w", err)
	}

	return &workflow, nil
}

// Validate checks if the workflow configuration is valid.
func (wc *WorkflowConfig) Validate() error {
	if wc.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(wc.Steps) == 0 && len(wc.Parallel) == 0 {
		return fmt.Errorf("workflow must have either steps or parallel configuration")
	}

	if len(wc.Steps) > 0 && len(wc.Parallel) > 0 {
		return fmt.Errorf("workflow cannot have both steps and parallel configuration")
	}

	// Validate steps
	steps := wc.Steps
	if len(wc.Parallel) > 0 {
		steps = wc.Parallel
	}

	for i, step := range steps {
		if step.Agent == "" {
			return fmt.Errorf("step %d: agent is required", i)
		}
		if step.Goal == "" {
			return fmt.Errorf("step %d: goal is required", i)
		}
	}

	return nil
}
