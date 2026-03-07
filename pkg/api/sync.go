package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/NikoSokratous/unagnt/internal/store"
	"github.com/NikoSokratous/unagnt/pkg/sync"
	"github.com/gorilla/mux"
)

// SyncAPI handles sync push/pull endpoints.
type SyncAPI struct {
	store *store.SQLite
}

// NewSyncAPI creates a sync API.
func NewSyncAPI(st *store.SQLite) *SyncAPI {
	return &SyncAPI{store: st}
}

// RegisterRoutes registers sync routes.
func (api *SyncAPI) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/v1/sync/push", api.Push).Methods("POST")
	router.HandleFunc("/v1/sync/pull", api.Pull).Methods("POST")
}

// Push handles POST /v1/sync/push
func (api *SyncAPI) Push(w http.ResponseWriter, r *http.Request) {
	var bundle sync.DeltaBundle
	if err := json.NewDecoder(r.Body).Decode(&bundle); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	adapter := &sync.StoreAdapter{Store: api.store}
	for _, rec := range bundle.Runs {
		run := &store.RunMeta{
			RunID:     rec.RunID,
			AgentName: rec.AgentName,
			Goal:      rec.Goal,
			State:     rec.State,
			StepCount: rec.StepCount,
			CreatedAt: rec.CreatedAt,
			UpdatedAt: rec.UpdatedAt,
		}
		if err := adapter.SaveRun(r.Context(), run); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
}

// Pull handles POST /v1/sync/pull
func (api *SyncAPI) Pull(w http.ResponseWriter, r *http.Request) {
	sinceStr := r.URL.Query().Get("since")
	var since time.Time
	if sinceStr != "" {
		since, _ = time.Parse(time.RFC3339, sinceStr)
	}

	adapter := &sync.StoreAdapter{Store: api.store}
	ls := sync.NewLocalSyncStore(adapter)
	bundle, err := ls.BuildBundle(r.Context(), since)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bundle)
}
