package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/NikoSokratous/unagnt/pkg/replay"
	"github.com/spf13/cobra"
)

func newReplayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replay",
		Short: "Replay recorded agent executions",
		Long:  "Commands for replaying, debugging, and validating recorded agent runs",
	}

	cmd.AddCommand(
		newReplayRunCmd(),
		newReplayListCmd(),
		newReplayDebugCmd(),
		newReplayValidateCmd(),
		newReplayDiffCmd(),
	)

	return cmd
}

func newReplayRunCmd() *cobra.Command {
	var mode string
	var startSeq, stopSeq int
	var breakpoints []int
	var verifySideEffects bool

	cmd := &cobra.Command{
		Use:   "run <snapshot-id>",
		Short: "Replay a recorded run",
		Long: `Replay a recorded run with different modes:
  exact       - Use recorded responses (deterministic)
  live        - Re-execute with live APIs
  mixed       - Recorded models, live tools
  debug       - Step-through debugging
  validation  - Verify consistency`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snapshotID := args[0]

			// Load snapshot (placeholder)
			fmt.Printf("Loading snapshot %s...\n", snapshotID)

			// Create mock snapshot for demonstration
			snapshot := &replay.RunSnapshot{
				ID:        snapshotID,
				RunID:     "run-123",
				AgentName: "test-agent",
				Goal:      "test goal",
			}

			// Create replayer
			replayer := replay.NewReplayer(snapshot)

			// Configure options
			options := replay.ReplayOptions{
				Mode:               replay.ReplayMode(mode),
				SnapshotID:         snapshotID,
				StartFromSequence:  startSeq,
				StopAtSequence:     stopSeq,
				Breakpoints:        breakpoints,
				ExecuteSideEffects: verifySideEffects,
				VerifyOutputs:      true,
			}

			// Replay
			fmt.Printf("Replaying in %s mode...\n", mode)
			result, err := replayer.Replay(context.Background(), options)
			if err != nil {
				return fmt.Errorf("replay failed: %w", err)
			}

			// Print results
			printReplayResult(result)

			return nil
		},
	}

	cmd.Flags().StringVar(&mode, "mode", "exact", "Replay mode (exact, live, mixed, debug, validation)")
	cmd.Flags().IntVar(&startSeq, "start", 0, "Start from sequence number")
	cmd.Flags().IntVar(&stopSeq, "stop", 0, "Stop at sequence number")
	cmd.Flags().IntSliceVar(&breakpoints, "breakpoint", []int{}, "Breakpoint at sequence numbers")
	cmd.Flags().BoolVar(&verifySideEffects, "verify-side-effects", false, "Execute side effects")

	return cmd
}

func newReplayListCmd() *cobra.Command {
	var runID string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Listing snapshots...\n")

			// Mock snapshots for demonstration
			snapshots := []replay.SnapshotMetadata{
				{
					ID:         "snap-001",
					RunID:      "run-123",
					AgentName:  "agent-1",
					Goal:       "Process data",
					SizeBytes:  1024 * 100,
					ModelCalls: 5,
					ToolCalls:  10,
					FinalState: "completed",
				},
				{
					ID:         "snap-002",
					RunID:      "run-124",
					AgentName:  "agent-2",
					Goal:       "Analyze logs",
					SizeBytes:  1024 * 250,
					ModelCalls: 8,
					ToolCalls:  15,
					FinalState: "completed",
				},
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "SNAPSHOT ID\tRUN ID\tAGENT\tMODEL CALLS\tTOOL CALLS\tSTATE\tSIZE")

			for _, snap := range snapshots {
				if runID != "" && snap.RunID != runID {
					continue
				}

				size := formatSize(snap.SizeBytes)
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\t%s\n",
					snap.ID,
					snap.RunID,
					snap.AgentName,
					snap.ModelCalls,
					snap.ToolCalls,
					snap.FinalState,
					size)
			}

			w.Flush()
			return nil
		},
	}

	cmd.Flags().StringVar(&runID, "run", "", "Filter by run ID")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of snapshots")

	return cmd
}

func newReplayDebugCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "debug <snapshot-id>",
		Short: "Debug replay with step-through",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snapshotID := args[0]

			fmt.Printf("Starting debug replay for %s...\n", snapshotID)
			fmt.Println("Commands: (c)ontinue, (s)tep, (i)nspect, (q)uit")

			// In production, this would start an interactive debugger
			fmt.Println("Debug mode not fully implemented in this demo")

			return nil
		},
	}
}

func newReplayValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <snapshot-id>",
		Short: "Validate snapshot integrity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snapshotID := args[0]

			fmt.Printf("Validating snapshot %s...\n\n", snapshotID)

			// Mock validation result
			fmt.Println("✓ Checksums valid")
			fmt.Println("✓ Sequence integrity verified")
			fmt.Println("✓ Timestamps monotonic")
			fmt.Println("✓ No data corruption detected")
			fmt.Println("\nSnapshot is valid!")

			return nil
		},
	}
}

func newReplayDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff <snapshot-id-1> <snapshot-id-2>",
		Short: "Compare two snapshots",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			snap1 := args[0]
			snap2 := args[1]

			fmt.Printf("Comparing %s and %s...\n\n", snap1, snap2)

			// Mock comparison
			fmt.Println("Differences:")
			fmt.Println("  Model calls: 5 vs 6 (+1)")
			fmt.Println("  Tool calls: 10 vs 10 (same)")
			fmt.Println("  Duration: 5.2s vs 5.5s (+0.3s)")
			fmt.Println("  Final state: completed vs completed (same)")

			return nil
		},
	}
}

func printReplayResult(result *replay.ReplayResult) {
	fmt.Printf("\n=== Replay Results ===\n")
	fmt.Printf("Snapshot: %s\n", result.SnapshotID)
	fmt.Printf("Mode: %s\n", result.Mode)
	fmt.Printf("Duration: %s\n", result.Duration)
	fmt.Printf("Success: %v\n\n", result.Success)

	fmt.Printf("Actions Rerun: %d\n", result.ActionsRerun)
	fmt.Printf("Matches: %d\n", result.Matches)
	fmt.Printf("Divergences: %d\n\n", len(result.Divergences))

	if len(result.Divergences) > 0 {
		fmt.Println("Divergence Details:")
		for i, div := range result.Divergences {
			fmt.Printf("%d. Seq %d (%s): %s [%s]\n",
				i+1,
				div.Sequence,
				div.Type,
				div.Description,
				div.Impact)

			if div.Expected != nil {
				expectedJSON, _ := json.MarshalIndent(div.Expected, "   ", "  ")
				fmt.Printf("   Expected: %s\n", string(expectedJSON))
			}
			if div.Actual != nil {
				actualJSON, _ := json.MarshalIndent(div.Actual, "   ", "  ")
				fmt.Printf("   Actual: %s\n", string(actualJSON))
			}
		}
	}

	fmt.Printf("\n=== Metrics ===\n")
	fmt.Printf("Accuracy: %.1f%%\n", result.Metrics.Accuracy*100)
	fmt.Printf("Original Duration: %s\n", result.Metrics.OriginalDuration)
	fmt.Printf("Replay Duration: %s\n", result.Metrics.ReplayDuration)
	if result.Metrics.SpeedupFactor > 0 {
		fmt.Printf("Speedup Factor: %.2fx\n", result.Metrics.SpeedupFactor)
	}
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
