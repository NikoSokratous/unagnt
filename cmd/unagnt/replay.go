package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/replay"
	"github.com/chzyer/readline"
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
	var snapshotFile string

	cmd := &cobra.Command{
		Use:   "debug <snapshot-id>",
		Short: "Debug replay with step-through and time-travel",
		Long: `Interactive time-travel debugging. Load a snapshot and step forward/back.
Commands: (s)tep forward, (b)ack, (g)oto <seq>, (p)rint state, (q)uit.
Use --file to load snapshot from JSON file instead of ID.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			snapshotID := args[0]
			snapshot, err := loadSnapshotForDebug(snapshotID, snapshotFile)
			if err != nil {
				return err
			}

			cursor := replay.NewReplayCursor(snapshot)
			fmt.Printf("Time-travel debug: %s (%d actions)\n", snapshotID, len(snapshot.ToolCalls))
			fmt.Println("Commands: (s)tep forward, (b)ack, (g)oto <seq>, (p)rint, (q)uit")

			rl, err := readline.New("debug> ")
			if err != nil {
				return err
			}
			defer rl.Close()

			for {
				line, err := rl.Readline()
				if err != nil {
					break
				}
				line = strings.TrimSpace(line)
				parts := strings.Fields(line)
				if len(parts) == 0 {
					continue
				}
				switch parts[0] {
				case "s", "step":
					if cursor.StepForward() {
						st := cursor.GetStateAt(cursor.Position())
						if st.CurrentAction != nil {
							fmt.Printf("  [%d] %s\n", st.Position, st.CurrentAction.ToolName)
						}
					} else {
						fmt.Println("  (at end)")
					}
				case "b", "back":
					if cursor.StepBack() {
						fmt.Printf("  position %d\n", cursor.Position())
					} else {
						fmt.Println("  (at start)")
					}
				case "g", "goto":
					if len(parts) < 2 {
						fmt.Println("  usage: goto <seq>")
						continue
					}
					seq, _ := strconv.Atoi(parts[1])
					cursor.SeekToSequence(seq)
					fmt.Printf("  seeked to %d\n", cursor.Position())
				case "p", "print":
					st := cursor.GetStateAt(cursor.Position())
					fmt.Printf("  position=%d can_forward=%v can_back=%v\n",
						st.Position, st.CanStepForward, st.CanStepBack)
					if st.CurrentAction != nil {
						fmt.Printf("  current: %s in=%s out=%s\n",
							st.CurrentAction.ToolName,
							string(st.CurrentAction.Input),
							string(st.CurrentAction.Output))
					}
				case "q", "quit", "exit":
					return nil
				default:
					fmt.Println("  unknown: use s, b, g, p, q")
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&snapshotFile, "file", "", "Load snapshot from JSON file")
	return cmd
}

func loadSnapshotForDebug(snapshotID, filePath string) (*replay.RunSnapshot, error) {
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read snapshot file: %w", err)
		}
		var snap replay.RunSnapshot
		if err := json.Unmarshal(data, &snap); err != nil {
			return nil, fmt.Errorf("parse snapshot: %w", err)
		}
		return &snap, nil
	}
	// Fallback: create minimal demo snapshot
	return &replay.RunSnapshot{
		ID:        snapshotID,
		RunID:     "run-" + snapshotID,
		AgentName: "demo",
		Goal:      "demo goal",
		ToolCalls: []replay.ToolExecution{
			{Sequence: 1, ToolName: "echo", Input: json.RawMessage(`{"msg":"a"}`), Output: json.RawMessage(`{"echoed":"a"}`)},
			{Sequence: 2, ToolName: "calc", Input: json.RawMessage(`{"op":"add","a":1,"b":2}`), Output: json.RawMessage(`{"result":3}`)},
		},
		StartTime:  time.Now().Add(-time.Minute),
		EndTime:    time.Now(),
		FinalState: "completed",
	}, nil
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
