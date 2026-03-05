package tool

import (
	"context"
	"database/sql"

	"github.com/NikoSokratous/unagnt/pkg/registry"
)

// Marketplace is a specialized client for the tool marketplace.
type Marketplace struct {
	registry *registry.Registry
}

// NewMarketplace creates a new tool marketplace client.
func NewMarketplace(db *sql.DB, serverURL, localPath string) *Marketplace {
	return &Marketplace{
		registry: registry.NewRegistry(db, serverURL, localPath),
	}
}

// SearchTools searches for tools in the marketplace.
func (m *Marketplace) SearchTools(ctx context.Context, query string, limit int) ([]registry.PluginMetadata, error) {
	filter := &registry.SearchFilter{
		Query:     query,
		Type:      registry.PluginTypeTool,
		Limit:     limit,
		SortBy:    "downloads",
		SortOrder: "desc",
	}

	result, err := m.registry.Search(ctx, filter)
	if err != nil {
		return nil, err
	}

	return result.Plugins, nil
}

// GetTool retrieves a specific tool.
func (m *Marketplace) GetTool(ctx context.Context, toolID string) (*registry.PluginMetadata, error) {
	return m.registry.Get(ctx, toolID)
}

// InstallTool installs a tool from the marketplace.
func (m *Marketplace) InstallTool(ctx context.Context, toolID string) error {
	// Placeholder: would download and install tool
	return nil
}

// ListInstalledTools lists locally installed tools.
func (m *Marketplace) ListInstalledTools() ([]registry.PluginMetadata, error) {
	return m.registry.ListInstalledPlugins()
}

// PublishTool publishes a tool to the marketplace.
func (m *Marketplace) PublishTool(ctx context.Context, manifest *registry.PluginManifest, artifact []byte) error {
	return m.registry.Register(ctx, manifest, artifact)
}
