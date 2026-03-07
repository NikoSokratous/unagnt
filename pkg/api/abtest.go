package api

import (
	"encoding/json"
	"net/http"

	"github.com/NikoSokratous/unagnt/pkg/abtest"
	"github.com/gorilla/mux"
)

// ABTestAPI handles A/B test configuration endpoints
type ABTestAPI struct {
	store *abtest.Store
}

// NewABTestAPI creates an A/B test API
func NewABTestAPI(store *abtest.Store) *ABTestAPI {
	return &ABTestAPI{store: store}
}

// RegisterRoutes registers A/B test routes
func (api *ABTestAPI) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/v1/ab-tests", api.List).Methods("GET")
	router.HandleFunc("/v1/ab-tests", api.Create).Methods("POST")
	router.HandleFunc("/v1/ab-tests/{id}", api.Patch).Methods("PATCH")
	router.HandleFunc("/v1/analytics/ab-tests", api.List).Methods("GET")
	router.HandleFunc("/v1/analytics/ab-tests/{id}/results", api.Results).Methods("GET")
}

// Create handles POST /v1/ab-tests
func (api *ABTestAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string  `json:"name"`
		ModelA       string  `json:"model_a"`
		ModelB       string  `json:"model_b"`
		TrafficSplit float64 `json:"traffic_split"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.ModelA == "" || req.ModelB == "" {
		http.Error(w, "model_a and model_b required", http.StatusBadRequest)
		return
	}
	if req.TrafficSplit < 0 || req.TrafficSplit > 1 {
		req.TrafficSplit = 0.5
	}

	t, err := api.store.Create(r.Context(), req.Name, req.ModelA, req.ModelB, req.TrafficSplit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

// List handles GET /v1/ab-tests and GET /v1/analytics/ab-tests
func (api *ABTestAPI) List(w http.ResponseWriter, r *http.Request) {
	tests, err := api.store.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ab_tests": tests})
}

// Patch handles PATCH /v1/ab-tests/{id}
func (api *ABTestAPI) Patch(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req struct {
		Active *bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Active != nil {
		if err := api.store.SetActive(r.Context(), id, *req.Active); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// Results handles GET /v1/analytics/ab-tests/{id}/results
func (api *ABTestAPI) Results(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	t, err := api.store.Get(r.Context(), id)
	if err != nil || t == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	// Placeholder results (would aggregate from ab_test_assignments + cost/tool metrics)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ab_test_id": id,
		"model_a":    t.ModelA,
		"model_b":    t.ModelB,
		"results": map[string]interface{}{
			"model_a": map[string]interface{}{"requests": 0, "latency_p99_ms": 0, "error_rate": 0},
			"model_b": map[string]interface{}{"requests": 0, "latency_p99_ms": 0, "error_rate": 0},
		},
	})
}
