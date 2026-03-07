package orchestrate

import (
	"os"
	"strings"

	"github.com/NikoSokratous/unagnt/internal/config"
)

func pickModelConfig(cfg *config.AgentConfig, goal string) config.ModelConfig {
	if cfg == nil {
		return config.ModelConfig{}
	}
	mr := cfg.ModelRouting
	if !mr.Enabled || len(mr.Candidates) == 0 {
		return cfg.Model
	}

	candidates := availableCandidates(mr.Candidates)
	if len(candidates) == 0 {
		return cfg.Model
	}

	switch strings.ToLower(strings.TrimSpace(mr.Strategy)) {
	case "cost":
		return bestByScore(candidates, costScore)
	case "latency":
		return bestByScore(candidates, latencyScore)
	case "capability":
		return bestByScore(candidates, capabilityScore)
	case "auto", "":
		return autoRoute(candidates, goal)
	default:
		return autoRoute(candidates, goal)
	}
}

func availableCandidates(in []config.ModelConfig) []config.ModelConfig {
	out := make([]config.ModelConfig, 0, len(in))
	for _, c := range in {
		if c.Provider == "" || c.Name == "" {
			continue
		}
		switch strings.ToLower(c.Provider) {
		case "openai":
			if os.Getenv("OPENAI_API_KEY") == "" {
				continue
			}
		case "anthropic":
			if os.Getenv("ANTHROPIC_API_KEY") == "" {
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func autoRoute(candidates []config.ModelConfig, goal string) config.ModelConfig {
	g := strings.ToLower(goal)
	if strings.Contains(g, "urgent") || strings.Contains(g, "quick") ||
		strings.Contains(g, "fast") || strings.Contains(g, "low latency") {
		return bestByScore(candidates, latencyScore)
	}
	if strings.Contains(g, "complex") || strings.Contains(g, "analyze") ||
		strings.Contains(g, "reason") || strings.Contains(g, "deep") {
		return bestByScore(candidates, capabilityScore)
	}
	return bestByScore(candidates, costScore)
}

func bestByScore(candidates []config.ModelConfig, scoreFn func(config.ModelConfig) int) config.ModelConfig {
	if len(candidates) == 0 {
		return config.ModelConfig{}
	}
	best := candidates[0]
	bestScore := scoreFn(best)
	for _, c := range candidates[1:] {
		s := scoreFn(c)
		if s > bestScore {
			best = c
			bestScore = s
		}
	}
	return best
}

func costScore(m config.ModelConfig) int {
	n := strings.ToLower(m.Name)
	switch {
	case strings.Contains(n, "mini"), strings.Contains(n, "haiku"), strings.Contains(n, "3.5"):
		return 100
	case strings.Contains(n, "sonnet"), strings.Contains(n, "4o"):
		return 60
	default:
		return 40
	}
}

func latencyScore(m config.ModelConfig) int {
	n := strings.ToLower(m.Name)
	switch {
	case strings.Contains(n, "mini"), strings.Contains(n, "haiku"), strings.Contains(n, "3.5"):
		return 100
	case strings.Contains(n, "sonnet"), strings.Contains(n, "4o"):
		return 70
	default:
		return 50
	}
}

func capabilityScore(m config.ModelConfig) int {
	n := strings.ToLower(m.Name)
	switch {
	case strings.Contains(n, "sonnet"), strings.Contains(n, "4o"), strings.Contains(n, "opus"):
		return 100
	case strings.Contains(n, "mini"), strings.Contains(n, "haiku"), strings.Contains(n, "3.5"):
		return 60
	default:
		return 50
	}
}
