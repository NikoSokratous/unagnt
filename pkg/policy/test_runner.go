package policy

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TestRunner runs policy tests from YAML files.
type TestRunner struct {
	simulator *Simulator
	store     *VersionStore
}

// NewTestRunner creates a new policy test runner.
func NewTestRunner(simulator *Simulator, store *VersionStore) *TestRunner {
	return &TestRunner{
		simulator: simulator,
		store:     store,
	}
}

// PolicyTest represents a policy test suite.
type PolicyTest struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Policy      string            `yaml:"policy"`
	Version     string            `yaml:"version,omitempty"`
	Tests       []TestCase        `yaml:"tests"`
	Setup       map[string]string `yaml:"setup,omitempty"`
}

// TestCase represents a single test case.
type TestCase struct {
	Name    string                 `yaml:"name"`
	Tool    string                 `yaml:"tool"`
	Input   map[string]interface{} `yaml:"input"`
	Context map[string]interface{} `yaml:"context,omitempty"`
	Expect  TestExpectation        `yaml:"expect"`
}

// TestExpectation defines expected test results.
type TestExpectation struct {
	Allowed        *bool    `yaml:"allowed,omitempty"`
	Denied         *bool    `yaml:"denied,omitempty"`
	Reason         string   `yaml:"reason,omitempty"`
	ReasonContains string   `yaml:"reasonContains,omitempty"`
	Alert          *bool    `yaml:"alert,omitempty"`
	MinRiskScore   *float64 `yaml:"minRiskScore,omitempty"`
	MaxRiskScore   *float64 `yaml:"maxRiskScore,omitempty"`
}

// TestResult represents the result of running tests.
type TestResult struct {
	TestFile   string           `json:"test_file"`
	TotalTests int              `json:"total_tests"`
	Passed     int              `json:"passed"`
	Failed     int              `json:"failed"`
	Skipped    int              `json:"skipped"`
	Duration   string           `json:"duration"`
	TestCases  []TestCaseResult `json:"test_cases"`
}

// TestCaseResult represents the result of a single test case.
type TestCaseResult struct {
	Name     string `json:"name"`
	Status   string `json:"status"` // "passed", "failed", "skipped"
	Message  string `json:"message,omitempty"`
	Expected string `json:"expected,omitempty"`
	Actual   string `json:"actual,omitempty"`
	Duration string `json:"duration"`
}

// RunTestFile runs tests from a YAML file.
func (r *TestRunner) RunTestFile(ctx context.Context, filename string) (*TestResult, error) {
	// Read test file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read test file: %w", err)
	}

	var test PolicyTest
	if err := yaml.Unmarshal(data, &test); err != nil {
		return nil, fmt.Errorf("parse test file: %w", err)
	}

	return r.RunTest(ctx, &test)
}

// RunTest runs a policy test suite.
func (r *TestRunner) RunTest(ctx context.Context, test *PolicyTest) (*TestResult, error) {
	result := &TestResult{
		TestFile:   test.Name,
		TotalTests: len(test.Tests),
		TestCases:  make([]TestCaseResult, 0, len(test.Tests)),
	}

	// Load policy
	policy, err := r.store.GetActiveVersion(ctx, test.Policy)
	if err != nil {
		return nil, fmt.Errorf("load policy: %w", err)
	}

	// Run each test case
	for _, testCase := range test.Tests {
		caseResult := r.runTestCase(ctx, policy, testCase)
		result.TestCases = append(result.TestCases, caseResult)

		switch caseResult.Status {
		case "passed":
			result.Passed++
		case "failed":
			result.Failed++
		case "skipped":
			result.Skipped++
		}
	}

	return result, nil
}

// runTestCase runs a single test case.
func (r *TestRunner) runTestCase(ctx context.Context, policy *PolicyVersion, testCase TestCase) TestCaseResult {
	result := TestCaseResult{
		Name:   testCase.Name,
		Status: "passed",
	}

	// Simulate the action
	simReq := SimulationRequest{
		PolicyName:    policy.PolicyName,
		PolicyVersion: policy.Version,
		Mode:          SimulationModeSimulation,
		Actions: []ActionToSimulate{
			{
				Sequence: 1,
				Tool:     testCase.Tool,
				Input:    testCase.Input,
				Context:  testCase.Context,
			},
		},
	}

	simResult, err := r.simulator.Simulate(ctx, simReq)
	if err != nil {
		result.Status = "failed"
		result.Message = fmt.Sprintf("simulation error: %v", err)
		return result
	}

	if len(simResult.Details) == 0 {
		result.Status = "failed"
		result.Message = "no simulation results"
		return result
	}

	actionResult := simResult.Details[0]

	// Check expectations
	if !r.checkExpectations(testCase.Expect, actionResult, &result) {
		result.Status = "failed"
	}

	return result
}

// checkExpectations validates test expectations.
func (r *TestRunner) checkExpectations(expect TestExpectation, actual ActionSimulationResult, result *TestCaseResult) bool {
	allMatch := true

	// Check allowed
	if expect.Allowed != nil {
		if actual.Allowed != *expect.Allowed {
			result.Expected = fmt.Sprintf("allowed=%v", *expect.Allowed)
			result.Actual = fmt.Sprintf("allowed=%v", actual.Allowed)
			result.Message = "allowed expectation mismatch"
			allMatch = false
		}
	}

	// Check denied
	if expect.Denied != nil {
		denied := !actual.Allowed
		if denied != *expect.Denied {
			result.Expected = fmt.Sprintf("denied=%v", *expect.Denied)
			result.Actual = fmt.Sprintf("denied=%v", denied)
			result.Message = "denied expectation mismatch"
			allMatch = false
		}
	}

	// Check reason contains
	if expect.ReasonContains != "" {
		if actual.DenyReason == "" || !contains(actual.DenyReason, expect.ReasonContains) {
			result.Expected = fmt.Sprintf("reason contains '%s'", expect.ReasonContains)
			result.Actual = fmt.Sprintf("reason='%s'", actual.DenyReason)
			result.Message = "deny reason expectation mismatch"
			allMatch = false
		}
	}

	// Check alert
	if expect.Alert != nil {
		if actual.WouldAlert != *expect.Alert {
			result.Expected = fmt.Sprintf("alert=%v", *expect.Alert)
			result.Actual = fmt.Sprintf("alert=%v", actual.WouldAlert)
			result.Message = "alert expectation mismatch"
			allMatch = false
		}
	}

	// Check risk score
	if expect.MinRiskScore != nil {
		if actual.RiskScore < *expect.MinRiskScore {
			result.Expected = fmt.Sprintf("riskScore >= %.2f", *expect.MinRiskScore)
			result.Actual = fmt.Sprintf("riskScore = %.2f", actual.RiskScore)
			result.Message = "risk score below minimum"
			allMatch = false
		}
	}

	if expect.MaxRiskScore != nil {
		if actual.RiskScore > *expect.MaxRiskScore {
			result.Expected = fmt.Sprintf("riskScore <= %.2f", *expect.MaxRiskScore)
			result.Actual = fmt.Sprintf("riskScore = %.2f", actual.RiskScore)
			result.Message = "risk score above maximum"
			allMatch = false
		}
	}

	return allMatch
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
