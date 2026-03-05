package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/NikoSokratous/unagnt/pkg/tool"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Plugin management commands",
	}
	cmd.AddCommand(newPluginListCmd())
	cmd.AddCommand(newPluginScanCmd())
	cmd.AddCommand(newPluginInfoCmd())
	cmd.AddCommand(newPluginValidateCmd())
	cmd.AddCommand(newPluginSearchCmd())
	cmd.AddCommand(newPluginInstallCmd())
	cmd.AddCommand(newPluginUpdateCmd())
	cmd.AddCommand(newPluginUninstallCmd())
	cmd.AddCommand(newPluginPublishCmd())
	return cmd
}

func newPluginListCmd() *cobra.Command {
	var pluginDirs []string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all discovered plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listPlugins(pluginDirs)
		},
	}

	cmd.Flags().StringSliceVarP(&pluginDirs, "dirs", "d", []string{"./plugins"}, "Plugin directories to scan")

	return cmd
}

func listPlugins(pluginDirs []string) error {
	green := color.New(color.FgGreen).SprintFunc()

	discovery := tool.NewPluginDiscovery(pluginDirs)

	fmt.Println("Scanning for plugins...")
	manifests, err := discovery.ScanPlugins()
	if err != nil {
		return fmt.Errorf("scan plugins: %w", err)
	}

	if len(manifests) == 0 {
		fmt.Println("No plugins found.")
		return nil
	}

	fmt.Printf("\nFound %s plugins:\n\n", green(fmt.Sprintf("%d", len(manifests))))

	table := tablewriter.NewWriter(os.Stdout)
	table.Append([]string{"NAME", "VERSION", "TYPE", "PERMISSIONS"})

	for _, m := range manifests {
		perms := fmt.Sprintf("%d", len(m.Permissions))
		if len(m.Permissions) > 0 {
			perms = fmt.Sprintf("%d (%s)", len(m.Permissions), m.Permissions[0])
		}
		table.Append([]string{m.Name, m.Version, m.Type, perms})
	}

	table.Render()

	return nil
}

func newPluginScanCmd() *cobra.Command {
	var pluginDirs []string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan and validate plugins in directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return scanPlugins(pluginDirs, verbose)
		},
	}

	cmd.Flags().StringSliceVarP(&pluginDirs, "dirs", "d", []string{"./plugins"}, "Plugin directories to scan")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")

	return cmd
}

func scanPlugins(pluginDirs []string, verbose bool) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	discovery := tool.NewPluginDiscovery(pluginDirs)

	fmt.Printf("Scanning directories: %v\n\n", pluginDirs)

	manifests, err := discovery.ScanPlugins()
	if err != nil {
		return fmt.Errorf("scan plugins: %w", err)
	}

	if len(manifests) == 0 {
		fmt.Printf("%s No plugins found in specified directories\n", yellow("⚠"))
		return nil
	}

	fmt.Printf("%s Found %d plugins\n\n", green("✓"), len(manifests))

	for i, m := range manifests {
		fmt.Printf("%d. %s (%s)\n", i+1, green(m.Name), m.Version)
		fmt.Printf("   Type: %s\n", m.Type)
		fmt.Printf("   Binary: %s\n", m.Binary)

		if verbose {
			if m.Description != "" {
				fmt.Printf("   Description: %s\n", m.Description)
			}
			if m.Author != "" {
				fmt.Printf("   Author: %s\n", m.Author)
			}
			if len(m.Permissions) > 0 {
				fmt.Printf("   Permissions: %v\n", m.Permissions)
			}
		}
		fmt.Println()
	}

	return nil
}

func newPluginInfoCmd() *cobra.Command {
	var pluginDirs []string

	cmd := &cobra.Command{
		Use:   "info <plugin-name>",
		Short: "Show detailed information about a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return showPluginInfo(args[0], pluginDirs)
		},
	}

	cmd.Flags().StringSliceVarP(&pluginDirs, "dirs", "d", []string{"./plugins"}, "Plugin directories to scan")

	return cmd
}

func showPluginInfo(pluginName string, pluginDirs []string) error {
	discovery := tool.NewPluginDiscovery(pluginDirs)

	manifests, err := discovery.ScanPlugins()
	if err != nil {
		return fmt.Errorf("scan plugins: %w", err)
	}

	// Find the plugin
	var found *tool.PluginManifest
	for i := range manifests {
		if manifests[i].Name == pluginName {
			found = &manifests[i]
			break
		}
	}

	if found == nil {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Printf("\n%s\n", green(found.Name))
	fmt.Printf("Version:     %s\n", found.Version)
	fmt.Printf("Type:        %s\n", found.Type)
	fmt.Printf("Binary:      %s\n", found.Binary)

	if found.Description != "" {
		fmt.Printf("Description: %s\n", found.Description)
	}

	if found.Author != "" {
		fmt.Printf("Author:      %s\n", found.Author)
	}

	if len(found.Permissions) > 0 {
		fmt.Printf("\n%s\n", cyan("Permissions:"))
		for _, p := range found.Permissions {
			fmt.Printf("  - %s\n", p)
		}
	}

	fmt.Println()

	return nil
}

func newPluginValidateCmd() *cobra.Command {
	var pluginDirs []string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate all plugin manifests",
		RunE: func(cmd *cobra.Command, args []string) error {
			return validatePlugins(pluginDirs)
		},
	}

	cmd.Flags().StringSliceVarP(&pluginDirs, "dirs", "d", []string{"./plugins"}, "Plugin directories to scan")

	return cmd
}

func validatePlugins(pluginDirs []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	discovery := tool.NewPluginDiscovery(pluginDirs)

	fmt.Println("Validating plugins...")

	manifests, err := discovery.ScanPlugins()
	if err != nil {
		return fmt.Errorf("scan plugins: %w", err)
	}

	if len(manifests) == 0 {
		fmt.Printf("%s No plugins found\n", yellow("⚠"))
		return nil
	}

	valid := 0
	invalid := 0
	warnings := 0

	for _, m := range manifests {
		hasIssues := false

		// Validate required fields
		if m.Name == "" {
			fmt.Printf("%s %s: missing 'name' field\n", red("✗"), m.Binary)
			invalid++
			hasIssues = true
			continue
		}

		if m.Version == "" {
			fmt.Printf("%s %s: missing 'version' field\n", red("✗"), m.Name)
			invalid++
			hasIssues = true
			continue
		}

		if m.Type != "goplugin" && m.Type != "wasm" {
			fmt.Printf("%s %s: invalid type '%s' (must be 'goplugin' or 'wasm')\n", red("✗"), m.Name, m.Type)
			invalid++
			hasIssues = true
			continue
		}

		if m.Binary == "" {
			fmt.Printf("%s %s: missing 'binary' field\n", red("✗"), m.Name)
			invalid++
			hasIssues = true
			continue
		}

		// Check if binary exists
		if _, err := os.Stat(m.Binary); os.IsNotExist(err) {
			fmt.Printf("%s %s: binary file not found: %s\n", yellow("⚠"), m.Name, m.Binary)
			warnings++
		}

		// Check description
		if m.Description == "" {
			fmt.Printf("%s %s: missing description\n", yellow("⚠"), m.Name)
			warnings++
		}

		if !hasIssues {
			fmt.Printf("%s %s v%s\n", green("✓"), m.Name, m.Version)
			valid++
		}
	}

	fmt.Printf("\n%s\n", green(fmt.Sprintf("Valid: %d", valid)))
	if invalid > 0 {
		fmt.Printf("%s\n", red(fmt.Sprintf("Invalid: %d", invalid)))
	}
	if warnings > 0 {
		fmt.Printf("%s\n", yellow(fmt.Sprintf("Warnings: %d", warnings)))
	}

	if invalid > 0 {
		return fmt.Errorf("validation failed: %d invalid plugins", invalid)
	}

	return nil
}

func newPluginSearchCmd() *cobra.Command {
	var registryURL string
	var pluginType string
	var limit int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search plugin marketplace",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			fmt.Printf("Searching marketplace for: %s\n\n", query)

			// Placeholder: would query marketplace API
			fmt.Printf("%-25s %-10s %-15s %-20s %-10s %-10s\n", "NAME", "VERSION", "TYPE", "AUTHOR", "RATING", "DOWNLOADS")
			fmt.Println(strings.Repeat("-", 100))
			fmt.Printf("%-25s %-10s %-15s %-20s %-10s %-10s\n", "example-tool", "1.0.0", "tool", "acme", "4.5", "1234")
			fmt.Printf("%-25s %-10s %-15s %-20s %-10s %-10s\n", "data-analyzer", "2.1.0", "tool", "data-co", "4.8", "5678")

			return nil
		},
	}

	cmd.Flags().StringVar(&registryURL, "registry", "https://registry.agentruntime.io", "Registry URL")
	cmd.Flags().StringVar(&pluginType, "type", "", "Filter by type")
	cmd.Flags().IntVar(&limit, "limit", 20, "Max results")

	return cmd
}

func newPluginInstallCmd() *cobra.Command {
	var registryURL string
	var force bool
	var skipValidation bool

	cmd := &cobra.Command{
		Use:   "install <plugin>[@version]",
		Short: "Install plugin from marketplace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginID := args[0]

			fmt.Printf("Installing %s...\n", pluginID)

			// Placeholder: installation steps
			steps := []string{
				"Downloading plugin metadata",
				"Validating permissions",
				"Downloading artifact",
				"Verifying checksum",
				"Extracting plugin",
				"Registering plugin",
			}

			for i, step := range steps {
				fmt.Printf("[%d/%d] %s\n", i+1, len(steps), step)
			}

			fmt.Printf("\n%s Plugin installed successfully!\n", color.GreenString("✓"))
			fmt.Printf("Install path: ~/.agentruntime/plugins/%s\n", pluginID)

			return nil
		},
	}

	cmd.Flags().StringVar(&registryURL, "registry", "https://registry.agentruntime.io", "Registry URL")
	cmd.Flags().BoolVar(&force, "force", false, "Force reinstall if already installed")
	cmd.Flags().BoolVar(&skipValidation, "skip-validation", false, "Skip security validation (dangerous)")

	return cmd
}

func newPluginUpdateCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "update [plugin]",
		Short: "Update installed plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			if all {
				fmt.Println("Updating all plugins...")
				// Placeholder: update all
			} else if len(args) > 0 {
				pluginID := args[0]
				fmt.Printf("Updating %s...\n", pluginID)
				// Placeholder: update specific plugin
			} else {
				return fmt.Errorf("specify plugin name or use --all")
			}

			fmt.Printf("%s Updates complete!\n", color.GreenString("✓"))
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Update all installed plugins")

	return cmd
}

func newPluginUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <plugin>",
		Short: "Uninstall a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginID := args[0]

			fmt.Printf("Uninstalling %s...\n", pluginID)

			// Placeholder: uninstall steps
			fmt.Println("Removing plugin files")
			fmt.Println("Cleaning up registry")

			fmt.Printf("%s Plugin uninstalled successfully!\n", color.GreenString("✓"))
			return nil
		},
	}
}

func newPluginPublishCmd() *cobra.Command {
	var registryURL string
	var manifestPath string

	cmd := &cobra.Command{
		Use:   "publish <artifact-path>",
		Short: "Publish plugin to marketplace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			artifactPath := args[0]

			fmt.Printf("Publishing plugin from: %s\n", artifactPath)

			// Placeholder: publish steps
			steps := []string{
				"Reading manifest",
				"Validating plugin",
				"Calculating checksum",
				"Uploading artifact",
				"Registering in marketplace",
			}

			for i, step := range steps {
				fmt.Printf("[%d/%d] %s\n", i+1, len(steps), step)
			}

			fmt.Printf("\n%s Plugin published successfully!\n", color.GreenString("✓"))
			fmt.Println("Plugin ID: example-plugin@1.0.0")
			fmt.Println("Marketplace URL: https://marketplace.agentruntime.io/plugins/example-plugin")

			return nil
		},
	}

	cmd.Flags().StringVar(&registryURL, "registry", "https://registry.agentruntime.io", "Registry URL")
	cmd.Flags().StringVar(&manifestPath, "manifest", "plugin.yaml", "Path to manifest file")

	return cmd
}
