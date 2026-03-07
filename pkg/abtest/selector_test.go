package abtest

import (
	"context"
	"testing"
)

func TestSelector_SelectModel(t *testing.T) {
	ctx := context.Background()
	s := NewSelector()

	t.Run("traffic_split_0", func(t *testing.T) {
		test := &ABTest{ModelA: "gpt-4", ModelB: "gpt-4-mini", TrafficSplit: 0, Active: true}
		chosen := s.SelectModel(ctx, test, "run1")
		if chosen != "gpt-4-mini" {
			t.Errorf("expected gpt-4-mini, got %s", chosen)
		}
	})

	t.Run("traffic_split_1", func(t *testing.T) {
		test := &ABTest{ModelA: "gpt-4", ModelB: "gpt-4-mini", TrafficSplit: 1, Active: true}
		chosen := s.SelectModel(ctx, test, "run1")
		if chosen != "gpt-4" {
			t.Errorf("expected gpt-4, got %s", chosen)
		}
	})

	t.Run("deterministic", func(t *testing.T) {
		test := &ABTest{ModelA: "gpt-4", ModelB: "gpt-4-mini", TrafficSplit: 0.5, Active: true}
		c1 := s.SelectModel(ctx, test, "run-123")
		c2 := s.SelectModel(ctx, test, "run-123")
		if c1 != c2 {
			t.Errorf("same runID should get same model: %s vs %s", c1, c2)
		}
	})
}
