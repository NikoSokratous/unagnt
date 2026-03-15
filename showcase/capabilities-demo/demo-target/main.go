// Demo target server - provides GET and POST /health on :8081
// Used by capabilities-demo so the agent has a localhost HTTP target.
package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		body, _ := io.ReadAll(r.Body)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "healthy",
			"method": r.Method,
			"received_body": string(body),
		})
	})
	log.Printf("Demo target listening on :8081 (GET/POST /health)")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
