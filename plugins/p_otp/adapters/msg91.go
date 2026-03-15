package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const Msg91BaseUrl = "https://control.msg91.com/api/v5"

// FlowRecipient represents the dynamic JSON payload for a MSG91 flow notification.
type FlowRecipient map[string]any

type Msg91Client struct {
	authKey string
}

func NewMsg91Client(authKey string) *Msg91Client {
	return &Msg91Client{
		authKey: authKey,
	}
}

// SendSMSFlow sends an SMS using the MSG91 Flow API.
func (c *Msg91Client) SendSMSFlow(templateID string, recipients []FlowRecipient, realTimeResponse bool) (map[string]any, error) {
	url := Msg91BaseUrl + "/flow"

	payload := map[string]any{
		"template_id": templateID,
		"short_url":   "1", // short_url is effectively a boolean (0 or 1) in MSG91
		"recipients":  recipients,
	}

	if realTimeResponse {
		payload["realTimeResponse"] = "1"
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("authkey", c.authKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("MSG91 API error: status %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]any
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w (body: %s)", err, string(bodyBytes))
	}

	return result, nil
}
