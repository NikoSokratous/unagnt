package mlops

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// PerformanceSample records a single LLM call outcome
type PerformanceSample struct {
	Provider  string
	Model     string
	LatencyMs int64
	Success   bool
	Timestamp time.Time
}

// Collector collects LLM call performance samples and persists snapshots
type Collector struct {
	db       *sql.DB
	samples  []PerformanceSample
	mu       sync.Mutex
	capacity int
}

// NewCollector creates a new performance collector
func NewCollector(db *sql.DB, capacity int) *Collector {
	if capacity <= 0 {
		capacity = 1000
	}
	return &Collector{
		db:       db,
		samples:  make([]PerformanceSample, 0, capacity),
		capacity: capacity,
	}
}

// Record records an LLM call outcome
func (c *Collector) Record(provider, model string, latencyMs int64, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.samples = append(c.samples, PerformanceSample{
		Provider:  provider,
		Model:     model,
		LatencyMs: latencyMs,
		Success:   success,
		Timestamp: time.Now(),
	})
	if len(c.samples) >= c.capacity {
		c.flushLocked()
	}
}

// Flush persists buffered samples as a snapshot
func (c *Collector) Flush(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.flushLockedWithContext(ctx)
}

func (c *Collector) flushLocked() {
	_ = c.flushLockedWithContext(context.Background())
}

func (c *Collector) flushLockedWithContext(ctx context.Context) error {
	if len(c.samples) == 0 {
		return nil
	}
	// Group by provider:model
	groups := make(map[string][]PerformanceSample)
	for _, s := range c.samples {
		key := fmt.Sprintf("%s:%s", s.Provider, s.Model)
		groups[key] = append(groups[key], s)
	}
	c.samples = c.samples[:0]

	for key, list := range groups {
		parts := strings.SplitN(key, ":", 2)
		provider, model := "unknown", "unknown"
		if len(parts) >= 1 && parts[0] != "" {
			provider = parts[0]
		}
		if len(parts) >= 2 && parts[1] != "" {
			model = parts[1]
		}

		// Compute percentiles and error rate
		latencies := make([]int64, 0, len(list))
		errors := 0
		for _, s := range list {
			latencies = append(latencies, s.LatencyMs)
			if !s.Success {
				errors++
			}
		}
		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

		p50 := percentile(latencies, 50)
		p95 := percentile(latencies, 95)
		p99 := percentile(latencies, 99)
		errorRate := float64(errors) / float64(len(list))
		throughput := float64(len(list)) / 60.0 // approximate per-minute

		query := `INSERT INTO model_performance_snapshots (model_id, version, provider, latency_p50_ms, latency_p95_ms, latency_p99_ms, error_rate, throughput, sample_count, timestamp)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err := c.db.ExecContext(ctx, query,
			model, "latest", provider,
			p50, p95, p99,
			errorRate, throughput,
			len(list), time.Now(),
		)
		if err != nil {
			return fmt.Errorf("insert snapshot: %w", err)
		}
	}
	return nil
}

func percentile(sorted []int64, p float64) int64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(float64(len(sorted))*p/100)) - 1
	if idx < 0 {
		idx = 0
	}
	return sorted[idx]
}
