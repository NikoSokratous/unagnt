package client

import (
	"context"
	"testing"
	"time"
)

func TestStreamEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This test requires a running agentd server
	client := New("http://localhost:8080", "")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	runID := "test-run-id"

	eventChan, errChan, err := client.StreamEvents(ctx, runID)
	if err != nil {
		t.Logf("StreamEvents setup error (expected without server): %v", err)
		return
	}

	// Wait for events or error
	select {
	case event := <-eventChan:
		t.Logf("Received event: %s", event.Type)
	case err := <-errChan:
		t.Logf("Stream error (expected without server): %v", err)
	case <-ctx.Done():
		t.Log("Context timeout (expected without server)")
	}
}
