package abtest

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// Store manages A/B test configurations
type Store struct {
	mu    sync.RWMutex
	tests map[string]*ABTest
}

// NewStore creates an in-memory A/B test store
func NewStore() *Store {
	return &Store{tests: make(map[string]*ABTest)}
}

// Create creates a new A/B test
func (s *Store) Create(ctx context.Context, name, modelA, modelB string, trafficSplit float64) (*ABTest, error) {
	id := uuid.New().String()
	t := &ABTest{
		ID:           id,
		ModelA:       modelA,
		ModelB:       modelB,
		TrafficSplit: trafficSplit,
		Active:       true,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tests[id] = t
	return t, nil
}

// Get returns an A/B test by ID
func (s *Store) Get(ctx context.Context, id string) (*ABTest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tests[id]
	if !ok {
		return nil, nil
	}
	return t, nil
}

// List returns all A/B tests
func (s *Store) List(ctx context.Context) ([]*ABTest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*ABTest, 0, len(s.tests))
	for _, t := range s.tests {
		out = append(out, t)
	}
	return out, nil
}

// SetActive activates or deactivates an A/B test
func (s *Store) SetActive(ctx context.Context, id string, active bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.tests[id]; ok {
		t.Active = active
	}
	return nil
}
