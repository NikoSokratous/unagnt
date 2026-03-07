package orchestrate

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NikoSokratous/unagnt/internal/store"
)

func TestHandleCreateRunRejectsNegativeHardeningFields(t *testing.T) {
	db := t.TempDir() + "/server-create-validation.db"
	st, err := store.NewSQLite(db)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer st.Close()

	s := NewServer("localhost:0", st, nil)
	body := `{"agent_name":"demo","goal":"x","max_retries":-1,"retry_backoff_ms":0,"timeout_ms":0}`
	req := httptest.NewRequest(http.MethodPost, "/v1/runs", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	s.handleCreateRun(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
}
