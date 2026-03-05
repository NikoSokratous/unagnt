package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TeamsConnector sends messages to Microsoft Teams via Incoming Webhook.
type TeamsConnector struct {
	webhookURL string
	client     *http.Client
}

// NewTeamsConnector creates a Teams connector.
func NewTeamsConnector(cfg *Config) *TeamsConnector {
	return &TeamsConnector{
		webhookURL: cfg.WebhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// teamsPayload is the Teams Adaptive Card / webhook format.
type teamsPayload struct {
	Type     string         `json:"@type"`
	Context  string         `json:"@context"`
	Summary  string         `json:"summary"`
	Sections []teamsSection `json:"sections"`
}

type teamsSection struct {
	ActivityTitle    string        `json:"activityTitle,omitempty"`
	ActivitySubtitle string        `json:"activitySubtitle,omitempty"`
	Text             string        `json:"text,omitempty"`
	PotentialAction  []teamsAction `json:"potentialAction,omitempty"`
}

type teamsAction struct {
	Type    string `json:"@type"`
	Name    string `json:"name"`
	Targets []struct {
		OS  string `json:"os"`
		URI string `json:"uri"`
	} `json:"targets"`
}

// Send posts a message to Teams.
func (t *TeamsConnector) Send(ctx context.Context, msg *Message) error {
	section := teamsSection{
		ActivityTitle: msg.Title,
		Text:          msg.Body,
	}
	for _, a := range msg.Actions {
		section.PotentialAction = append(section.PotentialAction, teamsAction{
			Type: "OpenUri",
			Name: a.Label,
			Targets: []struct {
				OS  string `json:"os"`
				URI string `json:"uri"`
			}{{OS: "default", URI: a.URL}},
		})
	}
	payload := teamsPayload{
		Type:     "MessageCard",
		Context:  "http://schema.org/extensions",
		Summary:  msg.Title,
		Sections: []teamsSection{section},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", t.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("teams webhook returned %d", resp.StatusCode)
	}
	return nil
}
