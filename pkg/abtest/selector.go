package abtest

import (
	"context"
	"hash/fnv"
	"sync"
)

// ABTest represents an active A/B test config
type ABTest struct {
	ID           string  `json:"id"`
	ModelA       string  `json:"model_a"`
	ModelB       string  `json:"model_b"`
	TrafficSplit float64 `json:"traffic_split"`
	Active       bool    `json:"active"`
}

// Selector selects model A or B based on traffic split
type Selector struct {
	mu sync.RWMutex
}

// NewSelector creates a selector
func NewSelector() *Selector {
	return &Selector{}
}

// SelectModel returns modelA or modelB based on deterministic traffic split for runID
func (s *Selector) SelectModel(ctx context.Context, test *ABTest, runID string) string {
	if test == nil || !test.Active {
		return ""
	}
	if test.TrafficSplit <= 0 {
		return test.ModelB
	}
	if test.TrafficSplit >= 1 {
		return test.ModelA
	}

	h := fnv.New64a()
	h.Write([]byte(runID))
	val := h.Sum64()
	pct := float64(val%10000) / 10000.0

	if pct < test.TrafficSplit {
		return test.ModelA
	}
	return test.ModelB
}

// ModelChosen returns "model_a" or "model_b" for assignment recording
func ModelChosen(test *ABTest, chosen string) string {
	if chosen == test.ModelA {
		return "model_a"
	}
	return "model_b"
}
