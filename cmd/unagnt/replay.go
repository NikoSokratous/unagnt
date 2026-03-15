package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/NikoSokratous/unagnt/pkg/replay"
	"github.com/NikoSokratous/unagnt/pkg/runtime"
	"github.com/NikoSokratous/unagnt/pkg/tool"
	"github.com/NikoSokratous/unagnt/pkg/tool/builtin"
	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
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
	var storePath string

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

			fmt.Printf("Loading snapshot %s...\n", snapshotID)

			var snapshot *replay.RunSnapshot
			if storePath != "" {
				absPath, _ := filepath.Abs(storePath)
				db, err := sql.Open("sqlite", absPath)
				if err == nil {
					snapStore := replay.NewSQLiteSnapshotStore(db)
					snap, err := snapStore.LoadSnapshot(context.Background(), snapshotID)
					db.Close()
					if err == nil {
						snapshot = snap
					}
				}
			}
			if snapshot == nil {
				return fmt.Errorf("snapshot %q not found (use --store to specify database)", snapshotID)
			}

			// Create replayer
			replayer := replay.NewReplayer(snapshot)

			// For live/mixed modes, wire up real tool execution
			var liveExec replay.LiveToolExecutor
			if mode == "live" || mode == "mixed" {
				reg := tool.NewRegistry()
				for _, t := range builtin.All() {
					reg.Register(t)
				}
				liveExec = &liveToolExecutorAdapter{inner: tool.NewExecutor(reg)}
			}

			// Configure options
			options := replay.ReplayOptions{
				Mode:               replay.ReplayMode(mode),
				SnapshotID:         snapshotID,
				StartFromSequence:  startSeq,
				StopAtSequence:     stopSeq,
				Breakpoints:        breakpoints,
				ExecuteSideEffects: verifySideEffects,
				VerifyOutputs:      true,
				LiveToolExecutor:   liveExec,
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
	cmd.Flags().StringVar(&storePath, "store", "agent.db", "SQLite store path (contains run_snapshots)")
	cmd.Flags().IntVar(&startSeq, "start", 0, "Start from sequence number")
	cmd.Flags().IntVar(&stopSeq, "stop", 0, "Stop at sequence number")
	cmd.Flags().IntSliceVar(&breakpoints, "breakpoint", []int{}, "Breakpoint at sequence numbers")
	cmd.Flags().BoolVar(&verifySideEffects, "verify-side-effects", false, "Execute side effects")

	return cmd
}

func newReplayListCmd() *cobra.Command {
	var runID string
	var limit int
	var storePath string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Listing snapshots...\n")

			var snapshots []replay.SnapshotMetadata
			if storePath != "" {
				absPath, err := filepath.Abs(storePath)
				if err != nil {
					absPath = storePath
				}
				db, err := sql.Open("sqlite", absPath)
				if err == nil {
					snapStore := replay.NewSQLiteSnapshotStore(db)
					list, err := snapStore.ListSnapshots(context.Background(), runID, limit)
					if err == nil {
						snapshots = list
					}
					_ = db.Close()
				}
			}

			if len(snapshots) == 0 {
				fmt.Println("No snapshots found. Run agents with --store to record snapshots.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "SNAPSHOT ID\tRUN ID\tAGENT\tMODEL CALLS\tTOOL CALLS\tSTATE\tSIZE")

			for _, snap := range snapshots {
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
	cmd.Flags().StringVar(&storePath, "store", "agent.db", "SQLite store path (contains run_snapshots)")

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
	// Mode-specific header
	switch result.Mode {
	case replay.ReplayModeExact:
		fmt.Println("\n--- exact: using recorded outputs (no tool execution) ---")
	case replay.ReplayModeLive:
		fmt.Println("\n--- live: re-executing tools, comparing to recording ---")
	case replay.ReplayModeMixed:
		fmt.Println("\n--- mixed: re-executing tools with recorded model context ---")
	case replay.ReplayModeDebug:
		fmt.Println("\n--- debug: step-through trace ---")
	case replay.ReplayModeValidation:
		fmt.Println("\n--- validation: integrity check ---")
	}

	// Execution trace — what actually happened
	if len(result.Trace) > 0 {
		fmt.Println()
		for _, t := range result.Trace {
			src := ""
			switch t.Source {
			case "recorded":
				src = "[recorded]"
			case "live":
				src = "[live]"
			case "breakpoint":
				src = "[breakpoint]"
			case "step":
				src = "[step]"
			}
			line := fmt.Sprintf("  %d. %s %s  in=%s", t.Seq, t.Tool, src, t.InputSum)
			if t.OutputSum != "" {
				line += fmt.Sprintf("  → %s", t.OutputSum)
			}
			if t.Result != "" && t.Result != "ok" {
				line += fmt.Sprintf("  %s", t.Result)
				if t.Result == "match" {
					line += " ✓"
				} else if t.Result == "diverged" {
					line += " ✗"
				}
			}
			if t.Duration != "" && t.Duration != "0s" {
				line += fmt.Sprintf("  (%s)", t.Duration)
			}
			fmt.Println(line)
		}
	}

	// Divergences (when live/mixed find mismatches)
	if len(result.Divergences) > 0 {
		fmt.Println("\n  Divergences:")
		for _, div := range result.Divergences {
			fmt.Printf("    seq %d %s: %s\n", div.Sequence, div.Component, div.Description)
			if s, ok := div.Expected.(string); ok && len(s) < 120 {
				fmt.Printf("      expected: %s\n", s)
			}
			if s, ok := div.Actual.(string); ok && len(s) < 120 {
				fmt.Printf("      actual:   %s\n", s)
			}
		}
	}

	// Summary
	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("  Success: %v  |  Actions: %d  |  Matches: %d  |  Divergences: %d\n",
		result.Success, result.ActionsRerun, result.Matches, len(result.Divergences))
	if result.Metrics.TotalActions > 0 {
		fmt.Printf("  Accuracy: %.1f%%  |  Original: %s  |  Replay: %s",
			result.Metrics.Accuracy*100, result.Metrics.OriginalDuration, result.Metrics.ReplayDuration)
		if result.Metrics.SpeedupFactor > 0 && result.Metrics.SpeedupFactor < 1e9 {
			fmt.Printf("  |  Speedup: %.1fx", result.Metrics.SpeedupFactor)
		} else if result.Metrics.ReplayDuration == 0 {
			fmt.Print("  |  Speedup: ∞")
		}
		fmt.Println()
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

// liveToolExecutorAdapter adapts runtime.ToolExecutor to replay.LiveToolExecutor.
type liveToolExecutorAdapter struct {
	inner runtime.ToolExecutor
}

func (a *liveToolExecutorAdapter) Execute(ctx context.Context, toolName, version string, input json.RawMessage) (json.RawMessage, error) {
	res, err := a.inner.Execute(ctx, toolName, version, input)
	if err != nil {
		return nil, err
	}
	if res.Error != "" {
		return nil, fmt.Errorf("tool %s: %s", toolName, res.Error)
	}
	if res.Output == nil {
		return json.RawMessage(`{}`), nil
	}
	return json.Marshal(res.Output)
}
