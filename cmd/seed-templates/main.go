package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/NikoSokratous/unagnt/pkg/registry"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	_ "modernc.org/sqlite"
)

// TemplateFile represents the structure of a template YAML file
type TemplateFile struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Version     string                 `yaml:"version"`
	Author      string                 `yaml:"author"`
	Category    string                 `yaml:"category"`
	Tags        []string               `yaml:"tags"`
	Icon        string                 `yaml:"icon"`
	Parameters  []TemplateParameter    `yaml:"parameters"`
	Workflow    map[string]interface{} `yaml:"workflow"`
}

type TemplateParameter struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"`
	Description string      `yaml:"description"`
	Required    bool        `yaml:"required"`
	Default     interface{} `yaml:"default,omitempty"`
}

func main() {
	// Check for database path argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: seed-templates <database-path>")
		fmt.Println("Example: seed-templates ./agentruntime.db")
		os.Exit(1)
	}

	dbPath := os.Args[1]

	// Open database (modernc.org/sqlite uses "sqlite" driver name)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create marketplace
	marketplace := registry.NewWorkflowMarketplace(db)

	// Load templates from directory
	templatesDir := "examples/workflows/templates"
	files, err := filepath.Glob(filepath.Join(templatesDir, "*.yaml"))
	if err != nil {
		fmt.Printf("Error reading templates directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d template files\n", len(files))

	ctx := context.Background()
	loaded := 0

	for _, file := range files {
		// Skip README
		if filepath.Base(file) == "README.yaml" {
			continue
		}

		fmt.Printf("Loading template: %s... ", filepath.Base(file))

		// Read YAML file
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}

		// Parse YAML
		var templateFile TemplateFile
		if err := yaml.Unmarshal(data, &templateFile); err != nil {
			fmt.Printf("ERROR parsing YAML: %v\n", err)
			continue
		}

		// Convert to registry format
		params := make([]registry.TemplateParameter, len(templateFile.Parameters))
		for i, p := range templateFile.Parameters {
			params[i] = registry.TemplateParameter{
				Name:        p.Name,
				Type:        p.Type,
				Description: p.Description,
				Required:    p.Required,
				Default:     p.Default,
			}
		}

		// Create template
		template := &registry.WorkflowTemplate{
			ID:           uuid.New().String(),
			Name:         templateFile.Name,
			Author:       templateFile.Author,
			Description:  templateFile.Description,
			Category:     templateFile.Category,
			Tags:         templateFile.Tags,
			TemplateYAML: string(data),
			Parameters:   params,
			Version:      templateFile.Version,
			License:      "MIT",
			Metadata: map[string]interface{}{
				"icon": templateFile.Icon,
			},
		}

		// Publish to marketplace
		if err := marketplace.Publish(ctx, template); err != nil {
			fmt.Printf("ERROR publishing: %v\n", err)
			continue
		}

		fmt.Printf("✓ OK\n")
		loaded++
	}

	fmt.Printf("\nSuccessfully loaded %d templates into the marketplace!\n", loaded)
}
