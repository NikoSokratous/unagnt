package tool

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/NikoSokratous/unagnt/pkg/policy"
	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

// PolicyExecutor wraps an executor with policy checks and approval gates.
type PolicyExecutor struct {
	Inner       runtime.ToolExecutor
	Policy      *policy.Engine
	RiskScorer  policy.RiskScorer
	Approval    *policy.ApprovalGate
	Environment string
}

// Execute implements runtime.ToolExecutor with policy enforcement.
func (e *PolicyExecutor) Execute(ctx context.Context, tool, version string, input json.RawMessage) (*runtime.ToolResult, error) {
	var inputMap map[string]any
	if err := json.Unmarshal(input, &inputMap); err != nil {
		inputMap = make(map[string]any)
	}

	riskScore := 0.3
	if e.RiskScorer != nil {
		riskScore = e.RiskScorer.Score(tool, inputMap)
	}

	ctx2 := policy.EvalContext{
		Tool:        tool,
		Input:       inputMap,
		Environment: e.Environment,
		RiskScore:   riskScore,
	}

	if e.Policy != nil {
		res := e.Policy.Check(ctx2)
		if res.Deny {
			return &runtime.ToolResult{
				ToolID: tool + "@" + version,
				Error:  "policy denied: " + res.Message,
			}, errors.New("policy denied")
		}
		if res.RequireApproval && e.Approval != nil {
			ok, err := e.Approval.RequestApproval(ctx, tool, inputMap, res.Approvers)
			if err != nil {
				return &runtime.ToolResult{ToolID: tool + "@" + version, Error: err.Error()}, err
			}
			if !ok {
				return &runtime.ToolResult{
					ToolID: tool + "@" + version,
					Error:  "approval denied",
				}, errors.New("approval denied")
			}
		}
	}

	return e.Inner.Execute(ctx, tool, version, input)
}
