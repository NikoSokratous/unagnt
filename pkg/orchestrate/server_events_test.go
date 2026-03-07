package orchestrate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NikoSokratous/unagnt/internal/store"
	"github.com/NikoSokratous/unagnt/pkg/observe"
)

func TestHandleGetRunEventsReturnsPersistedEvents(t *testing.T) {
	db := t.TempDir() + "/server-events.db"
	st, err := store.NewSQLite(db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	s := NewServer("localhost:0", st, nil)
	runID := "run-events-api-1"
	if err := st.SaveEvent(context.Background(), runID, &observe.Event{
		RunID:     runID,
		Timestamp: time.Now(),
		Type:      observe.EventInit,
		Agent:     "agent-a",
		Data:      map[string]any{"x": 1},
	}); err != nil {
		t.Fatalf("save event: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/runs/"+runID+"/events", nil)
	req.SetPathValue("id", runID)
	rr := httptest.NewRecorder()
	s.handleGetRunEvents(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var body struct {
		RunID  string          `json:"run_id"`
		Events []observe.Event `json:"events"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.RunID != runID || len(body.Events) == 0 {
		t.Fatalf("unexpected response: %+v", body)
	}
}
