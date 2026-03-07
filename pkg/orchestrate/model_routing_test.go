package orchestrate

import (
	"testing"

	"github.com/NikoSokratous/unagnt/internal/config"
)

func TestPickModelConfigFallsBackWhenDisabled(t *testing.T) {
	cfg := &config.AgentConfig{
		Model: config.ModelConfig{Provider: "openai", Name: "gpt-4o-mini"},
	}
	got := pickModelConfig(cfg, "any goal")
	if got.Provider != "openai" || got.Name != "gpt-4o-mini" {
		t.Fatalf("unexpected fallback model: %#v", got)
	}
}

func TestPickModelConfigCostStrategy(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "x")
	cfg := &config.AgentConfig{
		Model: config.ModelConfig{Provider: "openai", Name: "gpt-4o"},
		ModelRouting: config.ModelRoutingConfig{
			Enabled:  true,
			Strategy: "cost",
			Candidates: []config.ModelConfig{
				{Provider: "openai", Name: "gpt-4o"},
				{Provider: "openai", Name: "gpt-4o-mini"},
			},
		},
	}
	got := pickModelConfig(cfg, "summarize this")
	if got.Name != "gpt-4o-mini" {
		t.Fatalf("expected cheapest model gpt-4o-mini, got: %s", got.Name)
	}
}

func TestPickModelConfigCapabilityStrategy(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "x")
	cfg := &config.AgentConfig{
		Model: config.ModelConfig{Provider: "anthropic", Name: "claude-3-5-haiku-20241022"},
		ModelRouting: config.ModelRoutingConfig{
			Enabled:  true,
			Strategy: "capability",
			Candidates: []config.ModelConfig{
				{Provider: "anthropic", Name: "claude-3-5-haiku-20241022"},
				{Provider: "anthropic", Name: "claude-3-5-sonnet-20241022"},
			},
		},
	}
	got := pickModelConfig(cfg, "complex multi-step reasoning")
	if got.Name != "claude-3-5-sonnet-20241022" {
		t.Fatalf("expected highest capability model, got: %s", got.Name)
	}
}

func TestPickModelConfigAutoStrategy(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "x")
	cfg := &config.AgentConfig{
		Model: config.ModelConfig{Provider: "openai", Name: "gpt-4o-mini"},
		ModelRouting: config.ModelRoutingConfig{
			Enabled:  true,
			Strategy: "auto",
			Candidates: []config.ModelConfig{
				{Provider: "openai", Name: "gpt-4o"},
				{Provider: "openai", Name: "gpt-4o-mini"},
			},
		},
	}
	gotFast := pickModelConfig(cfg, "urgent quick answer")
	if gotFast.Name != "gpt-4o-mini" {
		t.Fatalf("expected low-latency model, got: %s", gotFast.Name)
	}

	gotComplex := pickModelConfig(cfg, "deep complex analysis")
	if gotComplex.Name != "gpt-4o" {
		t.Fatalf("expected capability model for complex goal, got: %s", gotComplex.Name)
	}
}

func TestPickModelConfigSkipsUnavailableProviderKeys(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	cfg := &config.AgentConfig{
		Model: config.ModelConfig{Provider: "ollama", Name: "llama3.2"},
		ModelRouting: config.ModelRoutingConfig{
			Enabled:  true,
			Strategy: "cost",
			Candidates: []config.ModelConfig{
				{Provider: "openai", Name: "gpt-4o-mini"},
				{Provider: "anthropic", Name: "claude-3-5-haiku-20241022"},
			},
		},
	}
	got := pickModelConfig(cfg, "normal goal")
	if got.Provider != "ollama" || got.Name != "llama3.2" {
		t.Fatalf("expected base model fallback when candidates unavailable, got: %#v", got)
	}
}
