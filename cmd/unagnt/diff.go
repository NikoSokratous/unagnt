package main

import (
	"context"
	"fmt"
	"os"

	"github.com/NikoSokratous/unagnt/internal/store"
	"github.com/NikoSokratous/unagnt/pkg/observe"
	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	var runIDA, runIDB, storePath string
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare two runs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) >= 2 {
				runIDA, runIDB = args[0], args[1]
			}
			return runDiff(runIDA, runIDB, storePath)
		},
	}
	cmd.Flags().StringVar(&runIDA, "run-id-a", "", "First run ID")
	cmd.Flags().StringVar(&runIDB, "run-id-b", "", "Second run ID")
	cmd.Flags().StringVar(&storePath, "store", "agent.db", "Path to SQLite store")
	_ = cmd.MarkFlagRequired("run-id-a")
	_ = cmd.MarkFlagRequired("run-id-b")
	return cmd
}

func runDiff(runIDA, runIDB, storePath string) error {
	s, err := store.NewSQLite(storePath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	ctx := context.Background()
	eventsA, err := s.GetEvents(ctx, runIDA)
	if err != nil {
		return fmt.Errorf("get events for %s: %w", runIDA, err)
	}
	eventsB, err := s.GetEvents(ctx, runIDB)
	if err != nil {
		return fmt.Errorf("get events for %s: %w", runIDB, err)
	}

	// Convert to []observe.Event for Diff
	var evtsA, evtsB []observe.Event
	for _, e := range eventsA {
		evtsA = append(evtsA, *e)
	}
	for _, e := range eventsB {
		evtsB = append(evtsB, *e)
	}

	result := observe.Diff(evtsA, evtsB)
	fmt.Fprintf(os.Stdout, "Run A: %s (%d steps)\n", result.RunA, result.StepsA)
	fmt.Fprintf(os.Stdout, "Run B: %s (%d steps)\n", result.RunB, result.StepsB)
	fmt.Fprintf(os.Stdout, "Result: %s (diverged at step %d)\n", result.Message, result.Diverged)
	return nil
}
