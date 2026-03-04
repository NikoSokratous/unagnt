package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/agentruntime/agentruntime/internal/store"
	"github.com/agentruntime/agentruntime/pkg/orchestrate"
)

func main() {
	addr := flag.String("addr", ":8080", "Listen address")
	storePath := flag.String("store", "agent.db", "SQLite store path")
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

	srv := orchestrate.NewServer(*addr, st, apiKeys)
	fmt.Printf("agentd listening on %s (store: %s)\n", *addr, *storePath)
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
