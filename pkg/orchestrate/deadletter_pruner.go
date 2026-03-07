package orchestrate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/NikoSokratous/unagnt/internal/store"
)

// DeadLetterRetentionConfig configures retention, pruning, and optional archival.
type DeadLetterRetentionConfig struct {
	RetentionHours    int    // hours to keep before prune (default 168 = 7 days)
	ArchiveBeforePrune bool  // if true, archive to dir before prune
	ArchiveDir        string // directory for archival (required if ArchiveBeforePrune)
	PruneInterval     time.Duration // how often to run prune (default 1h)
}

// DeadLetterPruner runs background pruning (and optional archival) of dead letters.
type DeadLetterPruner struct {
	store   *store.SQLite
	cfg     DeadLetterRetentionConfig
	stop    chan struct{}
	stopped sync.WaitGroup
}

// NewDeadLetterPruner creates a pruner. cfg.RetentionHours must be > 0 for pruning.
func NewDeadLetterPruner(st *store.SQLite, cfg DeadLetterRetentionConfig) *DeadLetterPruner {
	if cfg.RetentionHours <= 0 {
		cfg.RetentionHours = 168 // 7 days
	}
	if cfg.PruneInterval <= 0 {
		cfg.PruneInterval = time.Hour
	}
	return &DeadLetterPruner{
		store: st,
		cfg:   cfg,
		stop:  make(chan struct{}),
	}
}

// Start begins the background prune loop.
func (p *DeadLetterPruner) Start() {
	p.stopped.Add(1)
	go p.run()
}

// Stop stops the pruner and waits for the loop to exit.
func (p *DeadLetterPruner) Stop() {
	close(p.stop)
	p.stopped.Wait()
}

func (p *DeadLetterPruner) run() {
	defer p.stopped.Done()
	ticker := time.NewTicker(p.cfg.PruneInterval)
	defer ticker.Stop()
	for {
		select {
		case <-p.stop:
			return
		case <-ticker.C:
			p.pruneOnce()
		}
	}
}

func (p *DeadLetterPruner) pruneOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	olderThan := time.Now().Add(-time.Duration(p.cfg.RetentionHours) * time.Hour)

	if p.cfg.ArchiveBeforePrune && p.cfg.ArchiveDir != "" {
		// Archive older items (single batch) before pruning
		items, err := p.store.ListDeadLettersOlderThan(ctx, olderThan, 5000)
		if err == nil {
			for _, dl := range items {
				if err := p.archive(dl); err != nil {
					continue
				}
				DeadLettersArchived.Inc()
			}
		}
	}

	n, err := p.store.PruneDeadLetters(ctx, olderThan)
	if err != nil {
		return
	}
	if n > 0 {
		DeadLettersPruned.Add(float64(n))
	}
}

func (p *DeadLetterPruner) archive(dl store.DeadLetter) error {
	safe := dl.RunID
	for _, r := range dl.RunID {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		safe = ""
		break
	}
	if safe == "" {
		safe = "unknown"
	}
	if err := os.MkdirAll(p.cfg.ArchiveDir, 0755); err != nil {
		return err
	}
	name := fmt.Sprintf("dead_letter_%s_%s.json", safe, dl.FailedAt.Format("20060102T150405Z0700"))
	path := filepath.Join(p.cfg.ArchiveDir, name)
	data, err := json.MarshalIndent(dl, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
