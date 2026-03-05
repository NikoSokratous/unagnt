package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/NikoSokratous/unagnt/internal/config"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newWebhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "Webhook management commands",
	}
	cmd.AddCommand(newWebhookAddCmd())
	cmd.AddCommand(newWebhookListCmd())
	cmd.AddCommand(newWebhookTestCmd())
	return cmd
}

func newWebhookAddCmd() *cobra.Command {
	var path, agent, goalTemplate, secret, callbackURL, webhooksFile string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new webhook configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookAdd(path, agent, goalTemplate, secret, callbackURL, webhooksFile)
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "", "Webhook path (e.g., /webhooks/github/pr)")
	cmd.Flags().StringVarP(&agent, "agent", "a", "", "Agent name to trigger")
	cmd.Flags().StringVarP(&goalTemplate, "goal", "g", "", "Goal template (e.g., 'Review PR #{{.number}}')")
	cmd.Flags().StringVarP(&secret, "secret", "s", "", "Webhook secret for signature verification")
	cmd.Flags().StringVar(&callbackURL, "callback", "", "Callback URL (optional)")
	cmd.Flags().StringVarP(&webhooksFile, "file", "f", "webhooks.yaml", "Webhooks config file")

	cmd.MarkFlagRequired("path")
	cmd.MarkFlagRequired("agent")
	cmd.MarkFlagRequired("goal")

	return cmd
}

func runWebhookAdd(path, agent, goalTemplate, secret, callbackURL, webhooksFile string) error {
	green := color.New(color.FgGreen).SprintFunc()

	// Load existing webhooks or create new
	var webhooks config.Webhooks
	if data, err := os.ReadFile(webhooksFile); err == nil {
		yaml.Unmarshal(data, &webhooks)
	}

	// Create new webhook config
	newWebhook := config.WebhookConfig{
		Path:         path,
		Agent:        agent,
		GoalTemplate: goalTemplate,
		AuthSecret:   secret,
		CallbackURL:  callbackURL,
	}

	// Validate
	if err := newWebhook.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Add to collection
	webhooks.Webhooks = append(webhooks.Webhooks, newWebhook)

	// Write back to file
	data, err := yaml.Marshal(&webhooks)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}

	if err := os.WriteFile(webhooksFile, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	fmt.Printf("%s Webhook added to %s\n", green("✓"), webhooksFile)
	fmt.Printf("  Path: %s\n", path)
	fmt.Printf("  Agent: %s\n", agent)
	fmt.Printf("  Goal Template: %s\n", goalTemplate)

	return nil
}

func newWebhookListCmd() *cobra.Command {
	var webhooksFile string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhook configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookList(webhooksFile)
		},
	}

	cmd.Flags().StringVarP(&webhooksFile, "file", "f", "webhooks.yaml", "Webhooks config file")
	return cmd
}

func runWebhookList(webhooksFile string) error {
	// Load webhooks
	webhooks, err := config.LoadWebhooks(webhooksFile)
	if err != nil {
		return fmt.Errorf("load webhooks: %w", err)
	}

	if len(webhooks.Webhooks) == 0 {
		fmt.Println("No webhooks configured")
		return nil
	}

	// Display in table
	table := tablewriter.NewWriter(os.Stdout)
	table.Append([]string{"Path", "Agent", "Goal Template", "Has Secret", "Callback"})

	for _, wh := range webhooks.Webhooks {
		hasSecret := "No"
		if wh.AuthSecret != "" {
			hasSecret = "Yes"
		}

		callback := wh.CallbackURL
		if callback == "" {
			callback = "-"
		}

		table.Append([]string{
			wh.Path,
			wh.Agent,
			wh.GoalTemplate,
			hasSecret,
			callback,
		})
	}

	fmt.Printf("Webhooks (%d configured):\n", len(webhooks.Webhooks))
	table.Render()

	return nil
}

func newWebhookTestCmd() *cobra.Command {
	var path, payloadFile, webhooksFile string

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test webhook goal template rendering",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebhookTest(path, payloadFile, webhooksFile)
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "", "Webhook path to test")
	cmd.Flags().StringVarP(&payloadFile, "payload", "d", "", "JSON payload file")
	cmd.Flags().StringVarP(&webhooksFile, "file", "f", "webhooks.yaml", "Webhooks config file")

	cmd.MarkFlagRequired("path")
	cmd.MarkFlagRequired("payload")

	return cmd
}

func runWebhookTest(path, payloadFile, webhooksFile string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	// Load webhooks
	webhooks, err := config.LoadWebhooks(webhooksFile)
	if err != nil {
		return fmt.Errorf("load webhooks: %w", err)
	}

	// Find webhook by path
	var webhook *config.WebhookConfig
	for i := range webhooks.Webhooks {
		if webhooks.Webhooks[i].Path == path {
			webhook = &webhooks.Webhooks[i]
			break
		}
	}

	if webhook == nil {
		return fmt.Errorf("webhook not found: %s", path)
	}

	// Load payload
	payloadData, err := os.ReadFile(payloadFile)
	if err != nil {
		return fmt.Errorf("read payload: %w", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(payloadData, &payload); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	// Render goal template
	tmpl, err := template.New("test").Parse(webhook.GoalTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, payload); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	renderedGoal := buf.String()

	fmt.Printf("%s Webhook Test Results\n\n", green("✓"))
	fmt.Printf("Path: %s\n", path)
	fmt.Printf("Agent: %s\n", webhook.Agent)
	fmt.Printf("Template: %s\n", yellow(webhook.GoalTemplate))
	fmt.Printf("Rendered Goal: %s\n\n", green(renderedGoal))

	if webhook.CallbackURL != "" {
		// Test callback URL rendering too
		tmpl, err = template.New("callback").Parse(webhook.CallbackURL)
		if err == nil {
			buf.Reset()
			if err := tmpl.Execute(&buf, payload); err == nil {
				fmt.Printf("Callback URL: %s\n", buf.String())
			}
		}
	}

	return nil
}
