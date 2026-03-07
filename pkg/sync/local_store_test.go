package sync

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NikoSokratous/unagnt/internal/store"
	_ "modernc.org/sqlite"
)

func TestLocalSyncStore_BuildAndApplyBundle(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "sync1.db")
	st, err := store.NewSQLite(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	defer os.Remove(tmp)

	run := &store.RunMeta{
		RunID:     "run-1",
		AgentName: "a",
		Goal:      "g",
		State:     "completed",
		StepCount: 1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := st.SaveRun(context.Background(), run); err != nil {
		t.Fatal(err)
	}

	adapter := &StoreAdapter{Store: st}
	ls := NewLocalSyncStore(adapter)

	bundle, err := ls.BuildBundle(context.Background(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(bundle.Runs))
	}
	if bundle.Runs[0].RunID != "run-1" {
		t.Errorf("run_id want run-1, got %s", bundle.Runs[0].RunID)
	}

	// Apply to a fresh store
	tmp2 := filepath.Join(t.TempDir(), "sync2.db")
	st2, err := store.NewSQLite(tmp2)
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()
	defer os.Remove(tmp2)
	adapter2 := &StoreAdapter{Store: st2}
	ls2 := NewLocalSyncStore(adapter2)
	if err := ls2.ApplyBundle(context.Background(), bundle); err != nil {
		t.Fatal(err)
	}

	r, err := st2.GetRun(context.Background(), "run-1")
	if err != nil || r == nil {
		t.Fatal("run not found after apply")
	}
	if r.AgentName != "a" {
		t.Errorf("agent_name want a, got %s", r.AgentName)
	}
}
