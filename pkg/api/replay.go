package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/NikoSokratous/unagnt/pkg/replay"
	"github.com/gorilla/mux"
)

// ReplayStore abstracts snapshot persistence for the Replay API.
type ReplayStore interface {
	LoadSnapshot(ctx context.Context, id string) (*replay.RunSnapshot, error)
	ListSnapshots(ctx context.Context, runID string, limit int) ([]replay.SnapshotMetadata, error)
}

// ReplayAPI handles replay and time-travel debugging endpoints.
type ReplayAPI struct {
	store ReplayStore
}

// NewReplayAPI creates a replay API. Pass a store that can load and list snapshots.
func NewReplayAPI(store ReplayStore) *ReplayAPI {
	return &ReplayAPI{store: store}
}

// RegisterRoutes registers replay routes.
func (api *ReplayAPI) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/v1/replay/snapshots", api.ListSnapshots).Methods("GET")
	router.HandleFunc("/v1/replay/snapshots/{id}", api.GetSnapshot).Methods("GET")
	router.HandleFunc("/v1/replay/snapshots/{id}/seek", api.SeekSnapshot).Methods("POST")
}

// ListSnapshots handles GET /v1/replay/snapshots
func (api *ReplayAPI) ListSnapshots(w http.ResponseWriter, r *http.Request) {
	runID := r.URL.Query().Get("run_id")
	limit := 50
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = n
		}
	}

	list, err := api.store.ListSnapshots(r.Context(), runID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"snapshots": list})
}

// GetSnapshot handles GET /v1/replay/snapshots/{id}
func (api *ReplayAPI) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	snap, err := api.store.LoadSnapshot(r.Context(), id)
	if err != nil {
		http.Error(w, "snapshot not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snap)
}


// SeekSnapshot handles POST /v1/replay/snapshots/{id}/seek
func (api *ReplayAPI) SeekSnapshot(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	snap, err := api.store.LoadSnapshot(r.Context(), id)
	if err != nil {
		http.Error(w, "snapshot not found", http.StatusNotFound)
		return
	}

	var req struct {
		Sequence int `json:"sequence"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Sequence = 0
	}

	replayer := replay.NewReplayer(snap)
	st := replayer.GetStateAt(req.Sequence)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(st)
}
