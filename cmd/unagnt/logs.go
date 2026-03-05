package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	var runID, logFile string
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "View execution logs for a run",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogs(runID, logFile)
		},
	}
	cmd.Flags().StringVar(&runID, "run-id", "", "Filter by run ID (optional)")
	cmd.Flags().StringVar(&logFile, "log-file", "agent.log", "Path to log file")
	return cmd
}

func runLogs(runID, logFile string) error {
	f, err := os.Open(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "No log file at %s. Run with --log-file to write events.\n", logFile)
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var evt map[string]any
		if err := json.Unmarshal(line, &evt); err != nil {
			fmt.Fprintln(os.Stderr, string(line))
			continue
		}
		if runID != "" {
			if id, ok := evt["run_id"].(string); !ok || id != runID {
				continue
			}
		}
		pretty, _ := json.MarshalIndent(evt, "", "  ")
		fmt.Println(string(pretty))
	}
	return scanner.Err()
}
