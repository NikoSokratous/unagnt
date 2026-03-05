package orchestrate

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/template"
	"time"

	"github.com/NikoSokratous/unagnt/internal/config"
	"github.com/NikoSokratous/unagnt/internal/store"
	"github.com/google/uuid"
)

// WebhookHandler manages webhook endpoints.
type WebhookHandler struct {
	server    *Server
	webhooks  map[string]*config.WebhookConfig // path -> config
	templates map[string]*template.Template    // path -> parsed template
}

// NewWebhookHandler creates a webhook handler.
func NewWebhookHandler(server *Server, webhookConfigs []config.WebhookConfig) (*WebhookHandler, error) {
	h := &WebhookHandler{
		server:    server,
		webhooks:  make(map[string]*config.WebhookConfig),
		templates: make(map[string]*template.Template),
	}

	// Parse and register all webhooks
	for i := range webhookConfigs {
		wh := &webhookConfigs[i]

		// Parse goal template
		tmpl, err := template.New(wh.Path).Parse(wh.GoalTemplate)
		if err != nil {
			return nil, fmt.Errorf("parse goal template for %s: %w", wh.Path, err)
		}

		h.webhooks[wh.Path] = wh
		h.templates[wh.Path] = tmpl
	}

	return h, nil
}

// HandleWebhook processes an incoming webhook request.
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Find webhook config
	webhookConfig, ok := h.webhooks[path]
	if !ok {
		http.Error(w, "webhook not found", http.StatusNotFound)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify signature if secret is configured
	if webhookConfig.AuthSecret != "" {
		secret := webhookConfig.ResolveSecret()
		if secret != "" && !h.verifySignature(r, body, secret) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse JSON payload
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Render goal from template
	goal, err := h.renderGoal(webhookConfig.Path, payload)
	if err != nil {
		http.Error(w, fmt.Sprintf("render goal: %v", err), http.StatusInternalServerError)
		return
	}

	// Create run asynchronously
	runID := uuid.New().String()
	go h.executeWebhookRun(runID, webhookConfig, goal, payload)

	// Respond immediately
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"run_id": runID,
		"status": "accepted",
	})
}

// verifySignature verifies the HMAC-SHA256 signature.
func (h *WebhookHandler) verifySignature(r *http.Request, body []byte, secret string) bool {
	// Check common signature headers
	var signature string

	// GitHub style: X-Hub-Signature-256
	if sig := r.Header.Get("X-Hub-Signature-256"); sig != "" {
		signature = sig
		if len(signature) > 7 && signature[:7] == "sha256=" {
			signature = signature[7:]
		}
	} else if sig := r.Header.Get("X-Signature"); sig != "" {
		signature = sig
	} else if sig := r.Header.Get("X-Webhook-Signature"); sig != "" {
		signature = sig
	} else {
		// No signature header found
		return false
	}

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	// Compare
	return hmac.Equal([]byte(signature), []byte(expected))
}

// renderGoal renders the goal template with the payload data.
func (h *WebhookHandler) renderGoal(path string, payload map[string]any) (string, error) {
	tmpl := h.templates[path]

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, payload); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// executeWebhookRun executes the agent run triggered by the webhook.
func (h *WebhookHandler) executeWebhookRun(runID string, wh *config.WebhookConfig, goal string, payload map[string]any) {
	ctx := context.Background()

	// Store run metadata
	now := time.Now()
	meta := &store.RunMeta{
		RunID:     runID,
		AgentName: wh.Agent,
		Goal:      goal,
		State:     "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.server.store.SaveRun(ctx, meta); err != nil {
		// Log error but don't fail
		return
	}

	// Note: Actual agent execution would happen here
	// For now, just mark as completed
	// TODO: Integrate with runtime engine when available
	meta.State = "completed"
	meta.UpdatedAt = time.Now()
	h.server.store.SaveRun(ctx, meta)

	// Send callback if configured
	if wh.CallbackURL != "" {
		h.sendCallback(wh, runID, meta.State, payload)
	}
}

// sendCallback sends a callback to the configured URL.
func (h *WebhookHandler) sendCallback(wh *config.WebhookConfig, runID, state string, originalPayload map[string]any) {
	// Render callback URL template if it contains variables
	callbackURL := wh.CallbackURL
	tmpl, err := template.New("callback").Parse(callbackURL)
	if err == nil {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, originalPayload); err == nil {
			callbackURL = buf.String()
		}
	}

	// Prepare callback payload
	payload := map[string]any{
		"run_id": runID,
		"agent":  wh.Agent,
		"state":  state,
	}

	body, _ := json.Marshal(payload)

	// Send POST request with retries
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequest("POST", callbackURL, bytes.NewReader(body))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		// Add custom headers
		for k, v := range wh.Headers {
			req.Header.Set(k, v)
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Success
			return
		}

		// Retry on error
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}
}
