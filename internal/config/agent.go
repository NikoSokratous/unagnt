package config

import (
	"fmt"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

// AgentConfig is the schema for agent.yaml.
type AgentConfig struct {
	Name            string                `yaml:"name"`
	Version         string                `yaml:"version"`
	Description     string                `yaml:"description"`
	Model           ModelConfig           `yaml:"model"`
	ModelRouting    ModelRoutingConfig    `yaml:"model_routing"`
	Autonomy        int                   `yaml:"autonomy_level"`
	Tools           []ToolRef             `yaml:"tools"`
	Policy          string                `yaml:"policy"`
	Memory          MemoryConfig          `yaml:"memory"`
	MaxSteps        int                   `yaml:"max_steps"`
	Timeout         string                `yaml:"timeout"`
	Retry           RetryConfig           `yaml:"retry"`
	ContextAssembly ContextAssemblyConfig `yaml:"context_assembly"`
	MCPSources      []MCPSourceConfig     `yaml:"mcp_sources"`
	Budget          *BudgetConfig         `yaml:"budget"`
}

// BudgetConfig configures budget limits and alerts.
type BudgetConfig struct {
	Limit          float64 `yaml:"budget_limit" json:"budget_limit"`       // Hard cap in USD (0 = no limit)
	AlertThreshold float64 `yaml:"alert_threshold" json:"alert_threshold"` // Fraction to alert (e.g. 0.8 = 80%)
	AlertWebhook   string  `yaml:"alert_webhook" json:"alert_webhook"`     // URL to POST when alert fires
	Period         string  `yaml:"period" json:"period"`                   // daily, weekly, monthly
	TenantID       string  `yaml:"tenant_id" json:"tenant_id"`             // Optional scope
}

// MCPSourceConfig configures an MCP tool source (stdio or http).
type MCPSourceConfig struct {
	Type       string   `yaml:"type"`        // "stdio" or "http"
	Command    string   `yaml:"command"`     // for stdio
	Args       []string `yaml:"args"`        // for stdio
	URL        string   `yaml:"url"`         // for http
	ToolPrefix string   `yaml:"tool_prefix"` // optional prefix for tool names
}

// ModelConfig configures the LLM.
type ModelConfig struct {
	Provider    string  `yaml:"provider"`
	Name        string  `yaml:"name"`
	Temperature float64 `yaml:"temperature"`
}

// ModelRoutingConfig configures optional dynamic model routing.
type ModelRoutingConfig struct {
	Enabled    bool          `yaml:"enabled"`
	Strategy   string        `yaml:"strategy"` // auto, cost, latency, capability
	Candidates []ModelConfig `yaml:"candidates"`
}

// ToolRef references a tool by name and version.
type ToolRef struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// MemoryConfig configures memory tiers.
type MemoryConfig struct {
	Persistent bool `yaml:"persistent"`
	Semantic   bool `yaml:"semantic"`
}

// RetryConfig configures retry behavior.
type RetryConfig struct {
	MaxAttempts int    `yaml:"max_attempts"`
	Backoff     string `yaml:"backoff"`
}

// ContextAssemblyConfig configures context assembly.
type ContextAssemblyConfig struct {
	Enabled          bool                    `yaml:"enabled"`
	MaxContextTokens int                     `yaml:"max_context_tokens"`
	Parallel         bool                    `yaml:"parallel"`
	Embeddings       EmbeddingsConfig        `yaml:"embeddings"`
	Providers        []ContextProviderConfig `yaml:"providers"`
	Assembly         AssemblyConfig          `yaml:"assembly"`
}

// EmbeddingsConfig configures embedding generation.
type EmbeddingsConfig struct {
	Provider  string `yaml:"provider"`    // openai, local, disabled
	Model     string `yaml:"model"`       // model name
	APIKeyEnv string `yaml:"api_key_env"` // environment variable for API key
}

// ContextProviderConfig configures a context provider.
type ContextProviderConfig struct {
	Type     string         `yaml:"type"`
	Priority int            `yaml:"priority"`
	Enabled  bool           `yaml:"enabled"`
	Config   map[string]any `yaml:"config"`
}

// AssemblyConfig configures context assembly.
type AssemblyConfig struct {
	TokenBudget map[string]int `yaml:"token_budget"`
}

// Validate checks the config and returns an error if invalid.
func (c *AgentConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.Model.Provider == "" {
		return fmt.Errorf("model.provider is required")
	}
	if c.Model.Name == "" {
		return fmt.Errorf("model.name is required")
	}
	if c.MaxSteps <= 0 {
		c.MaxSteps = 50
	}
	if c.Autonomy < 0 || c.Autonomy > 4 {
		c.Autonomy = 2
	}
	for i := range c.Tools {
		if c.Tools[i].Version == "" {
			c.Tools[i].Version = "1"
		}
	}
	return nil
}

// AutonomyLevel returns the runtime AutonomyLevel.
func (c *AgentConfig) AutonomyLevel() runtime.AutonomyLevel {
	return runtime.AutonomyLevel(c.Autonomy)
}

// TimeoutDuration parses the timeout string.
func (c *AgentConfig) TimeoutDuration() time.Duration {
	if c.Timeout == "" {
		return 300 * time.Second
	}
	d, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return 300 * time.Second
	}
	return d
}
