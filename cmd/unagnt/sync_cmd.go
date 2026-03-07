package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/NikoSokratous/unagnt/internal/store"
	"github.com/NikoSokratous/unagnt/pkg/sync"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Local-first sync with remote",
		Long:  "Push local runs to server, pull remote runs. Requires unagntd with /v1/sync endpoints.",
	}

	cmd.AddCommand(newSyncPushCmd())
	cmd.AddCommand(newSyncPullCmd())
	cmd.AddCommand(newSyncStatusCmd())

	return cmd
}

func newSyncPushCmd() *cobra.Command {
	var url, apiKey, dbPath string

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push local runs to server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if url == "" {
				url = os.Getenv("UNAGNT_SERVER_URL")
			}
			if url == "" {
				url = "http://localhost:8080"
			}
			if apiKey == "" {
				apiKey = os.Getenv("AGENT_RUNTIME_API_KEY")
			}
			if dbPath == "" {
				dbPath = "agent.db"
			}

			st, err := store.NewSQLite(dbPath)
			if err != nil {
				return fmt.Errorf("open store: %w", err)
			}
			defer st.Close()

			adapter := &sync.StoreAdapter{Store: st}
			ls := sync.NewLocalSyncStore(adapter)
			bundle, err := ls.BuildBundle(context.Background(), time.Time{})
			if err != nil {
				return err
			}

			client := sync.NewClient(url, apiKey)
			if err := client.Push(context.Background(), bundle); err != nil {
				return fmt.Errorf("push: %w", err)
			}

			ls.SetLastPush(time.Now())
			fmt.Printf("Pushed %d runs to %s\n", len(bundle.Runs), url)
			return nil
		},
	}

	cmd.Flags().StringVar(&url, "url", "", "Server URL (default: UNAGNT_SERVER_URL or http://localhost:8080)")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key (default: AGENT_RUNTIME_API_KEY)")
	cmd.Flags().StringVar(&dbPath, "db", "", "Local store path (default: agent.db)")

	return cmd
}

func newSyncPullCmd() *cobra.Command {
	var url, apiKey, dbPath string

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull remote runs to local",
		RunE: func(cmd *cobra.Command, args []string) error {
			if url == "" {
				url = os.Getenv("UNAGNT_SERVER_URL")
			}
			if url == "" {
				url = "http://localhost:8080"
			}
			if apiKey == "" {
				apiKey = os.Getenv("AGENT_RUNTIME_API_KEY")
			}
			if dbPath == "" {
				dbPath = "agent.db"
			}

			st, err := store.NewSQLite(dbPath)
			if err != nil {
				return fmt.Errorf("open store: %w", err)
			}
			defer st.Close()

			adapter := &sync.StoreAdapter{Store: st}
			ls := sync.NewLocalSyncStore(adapter)
			client := sync.NewClient(url, apiKey)

			since := ls.LastPull()
			bundle, err := client.Pull(context.Background(), since)
			if err != nil {
				return fmt.Errorf("pull: %w", err)
			}

			if err := ls.ApplyBundle(context.Background(), bundle); err != nil {
				return err
			}

			fmt.Printf("Pulled %d runs from %s\n", len(bundle.Runs), url)
			return nil
		},
	}

	cmd.Flags().StringVar(&url, "url", "", "Server URL")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key")
	cmd.Flags().StringVar(&dbPath, "db", "", "Local store path")

	return cmd
}

func newSyncStatusCmd() *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show sync status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = "agent.db"
			}

			st, err := store.NewSQLite(dbPath)
			if err != nil {
				return fmt.Errorf("open store: %w", err)
			}
			defer st.Close()

			adapter := &sync.StoreAdapter{Store: st}
			ls := sync.NewLocalSyncStore(adapter)
			st2, err := ls.Status(context.Background())
			if err != nil {
				return err
			}

			fmt.Printf("Local runs: %d\n", st2.LocalRuns)
			if st2.LastPush != nil {
				fmt.Printf("Last push: %s\n", st2.LastPush.Format(time.RFC3339))
			} else {
				fmt.Println("Last push: never")
			}
			if st2.LastPull != nil {
				fmt.Printf("Last pull: %s\n", st2.LastPull.Format(time.RFC3339))
			} else {
				fmt.Println("Last pull: never")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "Local store path")

	return cmd
}
