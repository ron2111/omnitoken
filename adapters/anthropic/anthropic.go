// Package anthropic provides an API-backed Anthropic message token counter.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.anthropic.com"
const defaultVersion = "2023-06-01"

// Client calls Anthropic's server-side token counting API.
type Client struct {
	APIKey     string
	BaseURL    string
	Version    string
	HTTPClient *http.Client
}

// Message is a minimal Anthropic message shape.
type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// CountRequest is the request body for server-side message token counting.
type CountRequest struct {
	Model              string         `json:"model"`
	System             any            `json:"system,omitempty"`
	Messages           []Message      `json:"messages"`
	Tools              any            `json:"tools,omitempty"`
	Thinking           any            `json:"thinking,omitempty"`
	ToolChoice         any            `json:"tool_choice,omitempty"`
	Metadata           any            `json:"metadata,omitempty"`
	ExtraRequestFields map[string]any `json:"-"`
}

// CountResult is the token count returned by Anthropic.
type CountResult struct {
	InputTokens int `json:"input_tokens"`
}

// Usage is the token accounting block returned by Anthropic message responses.
type Usage struct {
	InputTokens              int             `json:"input_tokens,omitempty"`
	OutputTokens             int             `json:"output_tokens,omitempty"`
	CacheCreationInputTokens int             `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int             `json:"cache_read_input_tokens,omitempty"`
	ServerToolUse            map[string]int  `json:"server_tool_use,omitempty"`
	Raw                      json.RawMessage `json:"-"`
}

// CountMessageTokens calls Anthropic's messages/count_tokens endpoint.
func (c Client) CountMessageTokens(ctx context.Context, req CountRequest) (CountResult, error) {
	if c.APIKey == "" {
		return CountResult{}, errors.New("omnitoken anthropic: API key is required")
	}
	if req.Model == "" {
		return CountResult{}, errors.New("omnitoken anthropic: model is required")
	}
	if len(req.Messages) == 0 {
		return CountResult{}, errors.New("omnitoken anthropic: at least one message is required")
	}

	body, err := requestBody(req)
	if err != nil {
		return CountResult{}, err
	}
	baseURL := strings.TrimRight(c.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	version := c.Version
	if version == "" {
		version = defaultVersion
	}
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/messages/count_tokens", bytes.NewReader(body))
	if err != nil {
		return CountResult{}, err
	}
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", version)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return CountResult{}, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CountResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return CountResult{}, fmt.Errorf("omnitoken anthropic: count_tokens returned %s: %s", resp.Status, string(respBody))
	}
	var result CountResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return CountResult{}, err
	}
	return result, nil
}

// ParseUsage extracts usage from an Anthropic Messages response. If data is
// already a usage object, it is parsed directly.
func ParseUsage(data []byte) (Usage, error) {
	var wrapper struct {
		Usage *Usage `json:"usage"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return Usage{}, err
	}
	if wrapper.Usage != nil {
		wrapper.Usage.Raw = cloneRaw(data)
		return *wrapper.Usage, nil
	}
	var usage Usage
	if err := json.Unmarshal(data, &usage); err != nil {
		return Usage{}, err
	}
	usage.Raw = cloneRaw(data)
	return usage, nil
}

func requestBody(req CountRequest) ([]byte, error) {
	body := map[string]any{
		"model":    req.Model,
		"messages": req.Messages,
	}
	if req.System != nil {
		body["system"] = req.System
	}
	if req.Tools != nil {
		body["tools"] = req.Tools
	}
	if req.Thinking != nil {
		body["thinking"] = req.Thinking
	}
	if req.ToolChoice != nil {
		body["tool_choice"] = req.ToolChoice
	}
	if req.Metadata != nil {
		body["metadata"] = req.Metadata
	}
	for key, value := range req.ExtraRequestFields {
		body[key] = value
	}
	return json.Marshal(body)
}

func cloneRaw(data []byte) json.RawMessage {
	if len(data) == 0 {
		return nil
	}
	return append(json.RawMessage(nil), data...)
}
