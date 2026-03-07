package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/NikoSokratous/unagnt/internal/config"
	"github.com/NikoSokratous/unagnt/internal/store"
	"github.com/NikoSokratous/unagnt/pkg/orchestrate"
)

func main() {
	addr := flag.String("addr", ":8080", "Listen address")
	storePath := flag.String("store", "agent.db", "SQLite store path")
	webhooksPath := flag.String("webhooks", "", "Path to webhooks yaml config")
	flag.Parse()

	st, err := store.NewSQLite(*storePath)
	if err != nil {
		log.Fatalf("Failed to open store: %v", err)
	}
	defer st.Close()

	apiKeys := parseAPIKeys(os.Getenv("AGENT_RUNTIME_API_KEYS"))
	if len(apiKeys) == 0 {
		apiKeys = parseAPIKeys(os.Getenv("AGENTD_API_KEYS"))
	}

	serverCfg := orchestrate.ServerConfig{
		Addr:    *addr,
		Store:   st,
		APIKeys: apiKeys,
	}
	if *webhooksPath != "" {
		webhooks, err := config.LoadWebhooks(*webhooksPath)
		if err != nil {
			log.Fatalf("Failed to load webhooks: %v", err)
		}
		serverCfg.Webhooks = webhooks.Webhooks
	}

	srv := orchestrate.NewServerWithConfig(serverCfg)
	fmt.Printf("unagntd listening on %s (store: %s)\n", *addr, *storePath)
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func parseAPIKeys(v string) []string {
	if v == "" {
		return nil
	}
	var keys []string
	for _, k := range strings.Split(v, ",") {
		k = strings.TrimSpace(k)
		if k != "" {
			keys = append(keys, k)
		}
	}
	return keys
}
