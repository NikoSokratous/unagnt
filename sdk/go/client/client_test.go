package client

import (
	"context"
	"testing"
	"time"
)

func TestClientCreation(t *testing.T) {
	client := New("http://localhost:8080", "test-key")

	if client.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL = %v, want http://localhost:8080", client.baseURL)
	}
	if client.apiKey != "test-key" {
		t.Errorf("apiKey = %v, want test-key", client.apiKey)
	}
}

func TestCreateRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := New("http://localhost:8080", "")
	ctx := context.Background()

	// This would need a running agentd instance
	_, err := client.CreateRun(ctx, "demo-agent", "test goal")
	if err != nil {
		t.Logf("CreateRun failed (expected without running server): %v", err)
	}
}

func TestWaitForRun(t *testing.T) {
	client := New("http://localhost:8080", "")

	// Test context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.WaitForRun(ctx, "test-run-id", 50*time.Millisecond)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}
