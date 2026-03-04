package policy

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	// Create tables
	schema := `
		CREATE TABLE policy_versions (
			id TEXT PRIMARY KEY,
			policy_name TEXT NOT NULL,
			version TEXT NOT NULL,
			content BLOB NOT NULL,
			format TEXT DEFAULT 'yaml',
			author TEXT,
			changelog TEXT,
			effective_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL,
			supersedes TEXT,
			active BOOLEAN DEFAULT false,
			metadata JSON,
			UNIQUE(policy_name, version)
		);

		CREATE TABLE policy_audit (
			id TEXT PRIMARY KEY,
			timestamp TIMESTAMP NOT NULL,
			run_id TEXT,
			agent_name TEXT NOT NULL,
			policy_name TEXT NOT NULL,
			policy_version TEXT NOT NULL,
			action TEXT NOT NULL,
			tool TEXT NOT NULL,
			decision TEXT NOT NULL,
			risk_score REAL,
			deny_reason TEXT,
			context JSON,
			reviewed_by TEXT,
			reviewed_at TIMESTAMP
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestVersionStore(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tmpDir := t.TempDir()
	store, err := NewVersionStore(db, tmpDir)
	if err != nil {
		t.Fatalf("NewVersionStore: %v", err)
	}

	ctx := context.Background()

	// Test save version
	version := &PolicyVersion{
		PolicyName:  "test-policy",
		Version:     "1.0.0",
		Content:     []byte(`{"rules": []}`),
		Format:      "json",
		Author:      "test@example.com",
		Changelog:   "Initial version",
		EffectiveAt: time.Now(),
		Active:      true,
	}

	if err := store.SaveVersion(ctx, version); err != nil {
		t.Fatalf("SaveVersion: %v", err)
	}

	// Test get version
	retrieved, err := store.GetVersion(ctx, "test-policy", "1.0.0")
	if err != nil {
		t.Fatalf("GetVersion: %v", err)
	}

	if retrieved.PolicyName != version.PolicyName {
		t.Errorf("PolicyName: got %s, want %s", retrieved.PolicyName, version.PolicyName)
	}
	if retrieved.Version != version.Version {
		t.Errorf("Version: got %s, want %s", retrieved.Version, version.Version)
	}

	// Test get active version
	active, err := store.GetActiveVersion(ctx, "test-policy")
	if err != nil {
		t.Fatalf("GetActiveVersion: %v", err)
	}

	if !active.Active {
		t.Error("Expected active version")
	}

	// Test list versions
	versions, err := store.ListVersions(ctx, "test-policy")
	if err != nil {
		t.Fatalf("ListVersions: %v", err)
	}

	if len(versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(versions))
	}

	// Test save second version
	version2 := &PolicyVersion{
		PolicyName:  "test-policy",
		Version:     "1.1.0",
		Content:     []byte(`{"rules": [{"id": "1"}]}`),
		Format:      "json",
		Author:      "test@example.com",
		Changelog:   "Added rule 1",
		EffectiveAt: time.Now(),
		Active:      false,
	}

	if err := store.SaveVersion(ctx, version2); err != nil {
		t.Fatalf("SaveVersion 2: %v", err)
	}

	// Test activate version
	if err := store.SetActiveVersion(ctx, "test-policy", "1.1.0"); err != nil {
		t.Fatalf("SetActiveVersion: %v", err)
	}

	active, err = store.GetActiveVersion(ctx, "test-policy")
	if err != nil {
		t.Fatalf("GetActiveVersion after switch: %v", err)
	}

	if active.Version != "1.1.0" {
		t.Errorf("Active version: got %s, want 1.1.0", active.Version)
	}
}

func TestAuditLogger(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := NewAuditLogger(db)
	ctx := context.Background()

	// Test log entry
	log := &AuditLog{
		AgentName:     "test-agent",
		PolicyName:    "test-policy",
		PolicyVersion: "1.0.0",
		Action:        "read_file",
		Tool:          "file_reader",
		Decision:      "deny",
		RiskScore:     0.8,
		DenyReason:    "High risk operation",
		Context:       map[string]interface{}{"path": "/sensitive/data"},
	}

	if err := logger.Log(ctx, log); err != nil {
		t.Fatalf("Log: %v", err)
	}

	// Test query
	filter := AuditFilter{
		AgentName: "test-agent",
		Limit:     10,
	}

	logs, err := logger.Query(ctx, filter)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}

	if logs[0].Decision != "deny" {
		t.Errorf("Decision: got %s, want deny", logs[0].Decision)
	}

	// Test query by decision
	filter2 := AuditFilter{
		Decision: "deny",
		Limit:    10,
	}

	logs2, err := logger.Query(ctx, filter2)
	if err != nil {
		t.Fatalf("Query by decision: %v", err)
	}

	if len(logs2) != 1 {
		t.Errorf("Expected 1 denied log, got %d", len(logs2))
	}

	// Test mark reviewed
	if err := logger.MarkReviewed(ctx, logs[0].ID, "admin@example.com"); err != nil {
		t.Fatalf("MarkReviewed: %v", err)
	}

	// Verify reviewed
	logs3, err := logger.Query(ctx, filter)
	if err != nil {
		t.Fatalf("Query after review: %v", err)
	}

	if logs3[0].ReviewedBy != "admin@example.com" {
		t.Error("Expected log to be marked as reviewed")
	}
}

func TestSimulator(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tmpDir := t.TempDir()
	store, err := NewVersionStore(db, tmpDir)
	if err != nil {
		t.Fatalf("NewVersionStore: %v", err)
	}

	ctx := context.Background()

	// Create a test policy
	policyContent := []byte(`{
		"rules": [
			{
				"id": "1",
				"tool": "delete_file",
				"effect": "deny",
				"reason": "File deletion not allowed",
				"riskScore": 0.9
			}
		]
	}`)

	version := &PolicyVersion{
		PolicyName: "test-policy",
		Version:    "1.0.0",
		Content:    policyContent,
		Format:     "json",
		Author:     "test@example.com",
		Active:     true,
	}

	if err := store.SaveVersion(ctx, version); err != nil {
		t.Fatalf("SaveVersion: %v", err)
	}

	// Create simulator
	executor := NewExecutor()
	simulator := NewSimulator(store, executor)

	// Test simulation
	req := SimulationRequest{
		PolicyName:    "test-policy",
		PolicyVersion: "1.0.0",
		Mode:          SimulationModeSimulation,
		Actions: []ActionToSimulate{
			{
				Sequence: 1,
				Tool:     "delete_file",
				Input:    map[string]interface{}{"path": "/test/file.txt"},
			},
			{
				Sequence: 2,
				Tool:     "read_file",
				Input:    map[string]interface{}{"path": "/test/file.txt"},
			},
		},
	}

	result, err := simulator.Simulate(ctx, req)
	if err != nil {
		t.Fatalf("Simulate: %v", err)
	}

	if result.TotalActions != 2 {
		t.Errorf("TotalActions: got %d, want 2", result.TotalActions)
	}

	// First action should be denied (delete_file)
	if len(result.Details) < 1 {
		t.Fatal("No simulation details")
	}

	firstAction := result.Details[0]
	if firstAction.Allowed {
		t.Error("Expected delete_file to be denied")
	}
	if firstAction.DenyReason == "" {
		t.Error("Expected deny reason")
	}

	// Check summary
	if result.Summary.DenyRate == 0 {
		t.Error("Expected non-zero deny rate")
	}
}

func TestTestRunner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tmpDir := t.TempDir()
	store, err := NewVersionStore(db, tmpDir)
	if err != nil {
		t.Fatalf("NewVersionStore: %v", err)
	}

	ctx := context.Background()

	// Create a test policy
	policyContent := []byte(`{
		"rules": [
			{
				"id": "1",
				"tool": "sensitive_operation",
				"effect": "deny",
				"reason": "Sensitive operation blocked",
				"riskScore": 0.95
			}
		]
	}`)

	version := &PolicyVersion{
		PolicyName: "security-policy",
		Version:    "1.0.0",
		Content:    policyContent,
		Format:     "json",
		Author:     "security@example.com",
		Active:     true,
	}

	if err := store.SaveVersion(ctx, version); err != nil {
		t.Fatalf("SaveVersion: %v", err)
	}

	// Create test runner
	executor := NewExecutor()
	simulator := NewSimulator(store, executor)
	runner := NewTestRunner(simulator, store)

	// Create test case
	allowed := false
	minRisk := 0.9

	testSuite := &PolicyTest{
		Name:    "security-policy-tests",
		Policy:  "security-policy",
		Version: "1.0.0",
		Tests: []TestCase{
			{
				Name: "Block sensitive operation",
				Tool: "sensitive_operation",
				Input: map[string]interface{}{
					"action": "dangerous",
				},
				Expect: TestExpectation{
					Allowed:        &allowed,
					ReasonContains: "Sensitive",
					MinRiskScore:   &minRisk,
				},
			},
		},
	}

	// Run tests
	result, err := runner.RunTest(ctx, testSuite)
	if err != nil {
		t.Fatalf("RunTest: %v", err)
	}

	if result.TotalTests != 1 {
		t.Errorf("TotalTests: got %d, want 1", result.TotalTests)
	}

	if result.Failed > 0 {
		t.Errorf("Expected all tests to pass, but %d failed", result.Failed)
		for _, tc := range result.TestCases {
			if tc.Status == "failed" {
				t.Logf("Failed test: %s - %s", tc.Name, tc.Message)
			}
		}
	}
}

func TestAuditStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := NewAuditLogger(db)
	ctx := context.Background()

	// Create multiple audit logs
	now := time.Now()

	logs := []*AuditLog{
		{
			AgentName:     "agent1",
			PolicyName:    "policy1",
			PolicyVersion: "1.0.0",
			Action:        "action1",
			Tool:          "tool1",
			Decision:      "allow",
			RiskScore:     0.2,
			Timestamp:     now.Add(-1 * time.Hour),
		},
		{
			AgentName:     "agent1",
			PolicyName:    "policy1",
			PolicyVersion: "1.0.0",
			Action:        "action2",
			Tool:          "tool2",
			Decision:      "deny",
			RiskScore:     0.8,
			DenyReason:    "High risk",
			Timestamp:     now.Add(-2 * time.Hour),
		},
		{
			AgentName:     "agent2",
			PolicyName:    "policy1",
			PolicyVersion: "1.0.0",
			Action:        "action3",
			Tool:          "tool3",
			Decision:      "deny",
			RiskScore:     0.9,
			DenyReason:    "Prohibited",
			Timestamp:     now.Add(-3 * time.Hour),
		},
	}

	for _, log := range logs {
		if err := logger.Log(ctx, log); err != nil {
			t.Fatalf("Log: %v", err)
		}
	}

	// Get stats
	stats, err := logger.GetStats(ctx, now.Add(-24*time.Hour), now)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}

	if stats.Total != 3 {
		t.Errorf("Total: got %d, want 3", stats.Total)
	}

	if stats.Allowed != 1 {
		t.Errorf("Allowed: got %d, want 1", stats.Allowed)
	}

	if stats.Denied != 2 {
		t.Errorf("Denied: got %d, want 2", stats.Denied)
	}

	if stats.MaxRiskScore != 0.9 {
		t.Errorf("MaxRiskScore: got %.2f, want 0.9", stats.MaxRiskScore)
	}

	if len(stats.TopDenyReasons) != 2 {
		t.Errorf("TopDenyReasons: got %d, want 2", len(stats.TopDenyReasons))
	}
}

func TestExport(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logger := NewAuditLogger(db)
	ctx := context.Background()

	// Create audit log
	log := &AuditLog{
		AgentName:     "test-agent",
		PolicyName:    "test-policy",
		PolicyVersion: "1.0.0",
		Action:        "test_action",
		Tool:          "test_tool",
		Decision:      "allow",
		RiskScore:     0.5,
	}

	if err := logger.Log(ctx, log); err != nil {
		t.Fatalf("Log: %v", err)
	}

	// Test JSON export
	jsonData, err := logger.Export(ctx, AuditFilter{Limit: 10}, "json")
	if err != nil {
		t.Fatalf("Export JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected JSON data")
	}

	// Test CSV export
	csvData, err := logger.Export(ctx, AuditFilter{Limit: 10}, "csv")
	if err != nil {
		t.Fatalf("Export CSV: %v", err)
	}

	if len(csvData) == 0 {
		t.Error("Expected CSV data")
	}

	// Verify CSV has header
	csvStr := string(csvData)
	if !contains(csvStr, "ID,Timestamp") {
		t.Error("CSV should have header")
	}
}
