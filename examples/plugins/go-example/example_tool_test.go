package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestExampleToolInterface(t *testing.T) {
	tool := &ExampleTool{}

	if tool.Name() != "example_plugin" {
		t.Errorf("Expected name 'example_plugin', got '%s'", tool.Name())
	}

	if tool.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", tool.Version())
	}

	desc := tool.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
}

func TestExampleToolSchema(t *testing.T) {
	tool := &ExampleTool{}

	schema, err := tool.InputSchema()
	if err != nil {
		t.Fatalf("InputSchema failed: %v", err)
	}

	var schemaObj map[string]interface{}
	if err := json.Unmarshal(schema, &schemaObj); err != nil {
		t.Fatalf("Schema is not valid JSON: %v", err)
	}

	if schemaObj["type"] != "object" {
		t.Error("Schema type should be 'object'")
	}
}

func TestExampleToolPermissions(t *testing.T) {
	tool := &ExampleTool{}

	perms := tool.Permissions()
	if len(perms) == 0 {
		t.Error("Expected at least 1 permission")
	}

	if perms[0].Scope != "compute:cpu" {
		t.Errorf("Expected permission 'compute:cpu', got '%s'", perms[0].Scope)
	}
}

func TestExampleToolExecute(t *testing.T) {
	tool := &ExampleTool{}
	ctx := context.Background()

	tests := []struct {
		name      string
		input     string
		wantError bool
		checkFunc func(*testing.T, map[string]any)
	}{
		{
			name:      "simple message",
			input:     `{"message":"hello"}`,
			wantError: false,
			checkFunc: func(t *testing.T, output map[string]any) {
				if output["result"] != "hello" {
					t.Errorf("Expected result 'hello', got '%v'", output["result"])
				}
				if output["processed"] != true {
					t.Error("Expected processed=true")
				}
			},
		},
		{
			name:      "uppercase message",
			input:     `{"message":"hello","uppercase":true}`,
			wantError: false,
			checkFunc: func(t *testing.T, output map[string]any) {
				if output["result"] != "HELLO" {
					t.Errorf("Expected result 'HELLO', got '%v'", output["result"])
				}
				if output["uppercase"] != true {
					t.Error("Expected uppercase=true")
				}
			},
		},
		{
			name:      "invalid JSON",
			input:     `{invalid}`,
			wantError: true,
			checkFunc: nil,
		},
		{
			name:      "missing required field",
			input:     `{"uppercase":true}`,
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tool.Execute(ctx, json.RawMessage(tt.input))

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, output)
			}
		})
	}
}

func TestExampleToolConcurrency(t *testing.T) {
	tool := &ExampleTool{}
	ctx := context.Background()
	input := json.RawMessage(`{"message":"test"}`)

	// Run multiple concurrent executions
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_, err := tool.Execute(ctx, input)
			if err != nil {
				t.Errorf("Concurrent execution failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
