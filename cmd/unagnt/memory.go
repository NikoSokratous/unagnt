package main

import (
	"context"
	"fmt"
	"os"

	"github.com/NikoSokratous/unagnt/internal/store"
	"github.com/NikoSokratous/unagnt/pkg/memory"
	"github.com/spf13/cobra"
)

func newMemoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Memory operations (GDPR delete)",
	}
	cmd.AddCommand(newMemoryDeleteCmd())
	return cmd
}

func newMemoryDeleteCmd() *cobra.Command {
	var agentID, storePath string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete all non-log data for an agent (GDPR)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID == "" && len(args) > 0 {
				agentID = args[0]
			}
			return runMemoryDelete(agentID, storePath)
		},
	}
	cmd.Flags().StringVar(&agentID, "agent", "", "Agent ID to delete")
	cmd.Flags().StringVar(&storePath, "store", "agent.db", "Path to SQLite store")
	_ = cmd.MarkFlagRequired("agent")
	return cmd
}

func runMemoryDelete(agentID, storePath string) error {
	kv, err := store.NewSQLiteKV(storePath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer kv.Close()

	mgr := memory.NewManager(agentID, kv, nil)
	if err := mgr.DeleteAgent(context.Background()); err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Deleted all data for agent", agentID)
	return nil
}
