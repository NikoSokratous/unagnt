package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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
		Queue:   queueConfig(),
	}
	if dlRetention := deadLetterRetentionConfig(); dlRetention != nil {
		serverCfg.DeadLetterRetention = dlRetention
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

func queueConfig() orchestrate.QueueConfig {
	backend := os.Getenv("QUEUE_BACKEND")
	if backend == "" {
		backend = "memory"
	}
	size := 256
	if s := os.Getenv("QUEUE_SIZE"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			size = n
		}
	}
	return orchestrate.QueueConfig{
		Backend:   backend,
		RedisURL:  os.Getenv("QUEUE_REDIS_URL"),
		QueueSize: size,
	}
}

func deadLetterRetentionConfig() *orchestrate.DeadLetterRetentionConfig {
	hoursStr := os.Getenv("DEAD_LETTER_RETENTION_HOURS")
	if hoursStr == "" {
		return nil
	}
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		return nil
	}
	cfg := &orchestrate.DeadLetterRetentionConfig{
		RetentionHours: hours,
		PruneInterval:  time.Hour,
	}
	if dir := os.Getenv("DEAD_LETTER_ARCHIVE_DIR"); dir != "" {
		cfg.ArchiveBeforePrune = true
		cfg.ArchiveDir = dir
	}
	return cfg
}
