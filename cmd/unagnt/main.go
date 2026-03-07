package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "unagnt",
		Short: "CLI for the Agent Runtime",
	}

	root.AddCommand(newRunCmd())
	root.AddCommand(newLogsCmd())
	root.AddCommand(newReplayCmd())
	root.AddCommand(newDiffCmd())
	root.AddCommand(newMemoryCmd())
	root.AddCommand(newToolCmd())
	root.AddCommand(newPolicyCmd(nil)) // Policy commands (db initialized per command)
	root.AddCommand(newRiskCmd())      // Risk scoring commands
	root.AddCommand(newAgentCmd())
	root.AddCommand(newInitCmd())
	root.AddCommand(newScaffoldCmd())
	root.AddCommand(newDebugCmd())
	root.AddCommand(newWebhookCmd())
	root.AddCommand(newWorkflowCmd())
	root.AddCommand(newPluginCmd())
	root.AddCommand(newContextCmd()) // Context assembly commands
	root.AddCommand(newCostsCmd())
	root.AddCommand(newApprovalsCmd())
	root.AddCommand(newSyncCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
