// Approval server for HITL demo. Receives approval requests from agents,
// holds them until a human approves or denies via HTTP.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type PendingRequest struct {
	ID        string                 `json:"id"`
	Tool      string                 `json:"tool"`
	Input     map[string]interface{} `json:"input"`
	Approvers []string               `json:"approvers"`
	CreatedAt time.Time              `json:"created_at"`
	Decided   bool                   `json:"decided"`
	Approved  bool                   `json:"approved"`
}

var (
	mu      sync.Mutex
	pending = make(map[string]*PendingRequest)
	waiters = make(map[string]chan bool)
)

func main() {
	http.HandleFunc("POST /request", handleRequest)
	http.HandleFunc("GET /pending", handlePending)
	http.HandleFunc("POST /approve/{id}", handleApprove)
	http.HandleFunc("POST /deny/{id}", handleDeny)

	addr := ":9090"
	log.Printf("Approval server listening on %s", addr)
	log.Printf("  POST /request    - agent posts approval request (blocks until decided)")
	log.Printf("  GET /pending     - list pending requests")
	log.Printf("  POST /approve/:id - approve request")
	log.Printf("  POST /deny/:id    - deny request")
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Tool      string                 `json:"tool"`
		Input     map[string]interface{} `json:"input"`
		Approvers []string               `json:"approvers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	pr := &PendingRequest{
		ID:        id,
		Tool:      body.Tool,
		Input:     body.Input,
		Approvers: body.Approvers,
		CreatedAt: time.Now(),
	}

	mu.Lock()
	pending[id] = pr
	ch := make(chan bool, 1)
	waiters[id] = ch
	mu.Unlock()

	log.Printf("Approval request %s: tool=%s", id, body.Tool)

	// Block up to 5 minutes for human decision
	select {
	case approved := <-ch:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "approved": approved})
	case <-time.After(5 * time.Minute):
		mu.Lock()
		pr.Decided = true
		pr.Approved = false
		delete(waiters, id)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "approved": false, "timeout": true})
	}
}

func handlePending(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	list := make([]*PendingRequest, 0)
	for _, pr := range pending {
		if !pr.Decided {
			list = append(list, pr)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"pending": list})
}

func decide(id string, approved bool) bool {
	mu.Lock()
	defer mu.Unlock()
	pr := pending[id]
	if pr == nil || pr.Decided {
		return false
	}
	pr.Decided = true
	pr.Approved = approved
	if ch, ok := waiters[id]; ok {
		ch <- approved
		delete(waiters, id)
	}
	return true
}

func handleApprove(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if decide(id, true) {
		log.Printf("Approved %s", id)
		fmt.Fprintf(w, "OK approved %s", id)
	} else {
		http.NotFound(w, r)
	}
}

func handleDeny(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if decide(id, false) {
		log.Printf("Denied %s", id)
		fmt.Fprintf(w, "OK denied %s", id)
	} else {
		http.NotFound(w, r)
	}
}
