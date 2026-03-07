package api

import (
	"encoding/json"
	"net/http"

	"github.com/NikoSokratous/unagnt/pkg/policy"
	"github.com/gorilla/mux"
)

// ApprovalsAPI handles approval queue REST endpoints
type ApprovalsAPI struct {
	queue policy.ApprovalQueue
}

// NewApprovalsAPI creates an approvals API
func NewApprovalsAPI(queue policy.ApprovalQueue) *ApprovalsAPI {
	return &ApprovalsAPI{queue: queue}
}

// RegisterRoutes registers approval routes
func (api *ApprovalsAPI) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/v1/approvals/pending", api.ListPending).Methods("GET")
	router.HandleFunc("/v1/approvals/{id}", api.Get).Methods("GET")
	router.HandleFunc("/v1/approvals/{id}/approve", api.Approve).Methods("POST")
	router.HandleFunc("/v1/approvals/{id}/deny", api.Deny).Methods("POST")
}

// ListPending handles GET /v1/approvals/pending
func (api *ApprovalsAPI) ListPending(w http.ResponseWriter, r *http.Request) {
	list, err := api.queue.ListPending(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"pending": list})
}

// Get handles GET /v1/approvals/{id}
func (api *ApprovalsAPI) Get(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	req, err := api.queue.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if req == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

// Approve handles POST /v1/approvals/{id}/approve
func (api *ApprovalsAPI) Approve(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := api.queue.Approve(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "approved"})
}

// Deny handles POST /v1/approvals/{id}/deny
func (api *ApprovalsAPI) Deny(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := api.queue.Deny(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "denied"})
}
