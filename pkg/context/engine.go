package context

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/llm"
	"github.com/NikoSokratous/unagnt/pkg/runtime"
)

// ContextEngine orchestrates context assembly from multiple providers.
type ContextEngine struct {
	Providers []ContextProvider
	Assembler ContextAssembler
	Config    EngineConfig
	mu        sync.RWMutex
}

// EngineConfig configures the context engine.
type EngineConfig struct {
	MaxTokens     int
	EnableCache   bool
	CacheDuration time.Duration
	Parallel      bool // Fetch providers in parallel
}

// NewEngine creates a new context engine.
func NewEngine(config EngineConfig, assembler ContextAssembler) *ContextEngine {
	if config.MaxTokens == 0 {
		config.MaxTokens = 8000
	}
	if config.CacheDuration == 0 {
		config.CacheDuration = 30 * time.Second
	}
	return &ContextEngine{
		Providers: []ContextProvider{},
		Assembler: assembler,
		Config:    config,
	}
}

// AddProvider adds a context provider to the engine.
func (e *ContextEngine) AddProvider(provider ContextProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Providers = append(e.Providers, provider)
}

// FetchAll retrieves context from all providers.
func (e *ContextEngine) FetchAll(ctx context.Context, input ContextInput) ([]ContextFragment, error) {
	e.mu.RLock()
	providers := make([]ContextProvider, len(e.Providers))
	copy(providers, e.Providers)
	e.mu.RUnlock()

	if len(providers) == 0 {
		return []ContextFragment{}, nil
	}

	if e.Config.Parallel {
		return e.fetchParallel(ctx, input, providers)
	}
	return e.fetchSequential(ctx, input, providers)
}

// fetchSequential fetches context from providers sequentially.
func (e *ContextEngine) fetchSequential(ctx context.Context, input ContextInput, providers []ContextProvider) ([]ContextFragment, error) {
	fragments := []ContextFragment{}

	for _, provider := range providers {
		fragment, err := provider.Fetch(ctx, input)
		if err != nil {
			// Log error but continue with other providers
			continue
		}
		if fragment != nil {
			fragments = append(fragments, *fragment)
		}
	}

	// Sort by priority (lower number = higher priority)
	sort.Slice(fragments, func(i, j int) bool {
		return fragments[i].Priority < fragments[j].Priority
	})

	return fragments, nil
}

// fetchParallel fetches context from providers in parallel.
func (e *ContextEngine) fetchParallel(ctx context.Context, input ContextInput, providers []ContextProvider) ([]ContextFragment, error) {
	type result struct {
		fragment *ContextFragment
		err      error
	}

	results := make(chan result, len(providers))

	for _, provider := range providers {
		go func(p ContextProvider) {
			fragment, err := p.Fetch(ctx, input)
			results <- result{fragment: fragment, err: err}
		}(provider)
	}

	fragments := []ContextFragment{}
	for i := 0; i < len(providers); i++ {
		res := <-results
		if res.err == nil && res.fragment != nil {
			fragments = append(fragments, *res.fragment)
		}
	}

	// Sort by priority
	sort.Slice(fragments, func(i, j int) bool {
		return fragments[i].Priority < fragments[j].Priority
	})

	return fragments, nil
}

// Assemble assembles context fragments into messages.
func (e *ContextEngine) Assemble(fragments []ContextFragment, config AssemblyConfig) ([]llm.Message, error) {
	if e.Assembler == nil {
		return nil, fmt.Errorf("no assembler configured")
	}

	if config.MaxTokens == 0 {
		config.MaxTokens = e.Config.MaxTokens
	}

	return e.Assembler.Assemble(fragments, config)
}

// ContextProvider retrieves a specific type of context.
type ContextProvider interface {
	Name() string
	Priority() int // Lower = higher priority
	Fetch(ctx context.Context, input ContextInput) (*ContextFragment, error)
}

// ContextInput provides the input for context assembly.
type ContextInput struct {
	StepInput runtime.StepInput
	Tools     []llm.ToolInfo
	Metadata  map[string]any
}

// ContextFragment is a piece of assembled context.
type ContextFragment struct {
	ProviderName string
	Type         FragmentType
	Content      string
	Priority     int
	TokenCount   int
	Metadata     map[string]any
	Timestamp    time.Time
}

// FragmentType categorizes context fragments.
type FragmentType string

const (
	FragmentTypeSystemPrompt  FragmentType = "system_prompt"
	FragmentTypeMemory        FragmentType = "memory"
	FragmentTypePolicy        FragmentType = "policy"
	FragmentTypeWorkflowState FragmentType = "workflow_state"
	FragmentTypeToolOutput    FragmentType = "tool_output"
	FragmentTypeKnowledge     FragmentType = "knowledge"
	FragmentTypeUserGoal      FragmentType = "user_goal"
)

// ContextAssembler merges fragments into final messages.
type ContextAssembler interface {
	Assemble(fragments []ContextFragment, config AssemblyConfig) ([]llm.Message, error)
}

// AssemblyConfig configures the assembly process.
type AssemblyConfig struct {
	MaxTokens    int
	TokenBudget  map[FragmentType]int
	IncludeDebug bool
}
