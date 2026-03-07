package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func newApprovalsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approvals",
		Short: "Manage human approval requests",
		Long:  "List pending approvals, approve or deny requests (requires server with /v1/approvals API)",
	}
	cmd.AddCommand(newApprovalsListCmd())
	cmd.AddCommand(newApprovalsApproveCmd())
	cmd.AddCommand(newApprovalsDenyCmd())
	return cmd
}

func newApprovalsListCmd() *cobra.Command {
	var baseURL string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pending approval requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			if baseURL == "" {
				baseURL = "http://localhost:8080"
			}
			resp, err := http.Get(baseURL + "/v1/approvals/pending")
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("server returned %d", resp.StatusCode)
			}
			var out struct {
				Pending []struct {
					ID        string   `json:"id"`
					Tool      string   `json:"tool"`
					Approvers []string `json:"approvers"`
					RunID     string   `json:"run_id"`
					CreatedAt string   `json:"created_at"`
				} `json:"pending"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				return err
			}
			if len(out.Pending) == 0 {
				fmt.Println("No pending approvals")
				return nil
			}
			for _, p := range out.Pending {
				fmt.Printf("%s  %s  run=%s  approvers=%v\n", p.ID, p.Tool, p.RunID, p.Approvers)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&baseURL, "url", "http://localhost:8080", "API base URL")
	return cmd
}

func newApprovalsApproveCmd() *cobra.Command {
	var baseURL string
	cmd := &cobra.Command{
		Use:   "approve [id]",
		Short: "Approve an approval request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if baseURL == "" {
				baseURL = "http://localhost:8080"
			}
			req, _ := http.NewRequest("POST", baseURL+"/v1/approvals/"+args[0]+"/approve", nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("server returned %d", resp.StatusCode)
			}
			fmt.Println("Approved")
			return nil
		},
	}
	cmd.Flags().StringVar(&baseURL, "url", "http://localhost:8080", "API base URL")
	return cmd
}

func newApprovalsDenyCmd() *cobra.Command {
	var baseURL string
	cmd := &cobra.Command{
		Use:   "deny [id]",
		Short: "Deny an approval request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if baseURL == "" {
				baseURL = "http://localhost:8080"
			}
			req, _ := http.NewRequest("POST", baseURL+"/v1/approvals/"+args[0]+"/deny", nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("server returned %d", resp.StatusCode)
			}
			fmt.Println("Denied")
			return nil
		},
	}
	cmd.Flags().StringVar(&baseURL, "url", "http://localhost:8080", "API base URL")
	return cmd
}
