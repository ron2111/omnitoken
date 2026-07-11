package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultDeveloperBaseURL = "https://generativelanguage.googleapis.com"
	defaultVertexVersion    = "v1"
)

// Client calls Gemini/Google server-side token accounting APIs.
//
// Use APIKey for the Gemini Developer API. Use BearerToken with Project and
// Location for Vertex AI publisher model endpoints.
type Client struct {
	APIKey      string
	BearerToken string
	BaseURL     string
	APIVersion  string
	Project     string
	Location    string
	HTTPClient  *http.Client
}

// Content is the Gemini generate/count content shape.
type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts,omitempty"`
}

// Part is a Gemini content part. Exactly one data field is normally set.
type Part struct {
	Text                string    `json:"text,omitempty"`
	InlineData          *Blob     `json:"inlineData,omitempty"`
	FileData            *FileData `json:"fileData,omitempty"`
	FunctionCall        any       `json:"functionCall,omitempty"`
	FunctionResponse    any       `json:"functionResponse,omitempty"`
	ExecutableCode      any       `json:"executableCode,omitempty"`
	CodeExecutionResult any       `json:"codeExecutionResult,omitempty"`
	VideoMetadata       any       `json:"videoMetadata,omitempty"`
	MediaResolution     any       `json:"mediaResolution,omitempty"`
	Thought             bool      `json:"thought,omitempty"`
	ThoughtSignature    string    `json:"thoughtSignature,omitempty"`
}

// Blob is inline multimodal data encoded as base64 JSON.
type Blob struct {
	MIMEType string `json:"mimeType,omitempty"`
	Data     string `json:"data,omitempty"`
}

// FileData points to provider-accessible media, documents, or uploaded files.
type FileData struct {
	MIMEType string `json:"mimeType,omitempty"`
	FileURI  string `json:"fileUri,omitempty"`
}

// InlineDataFromBytes base64-encodes data for an inlineData part.
func InlineDataFromBytes(mimeType string, data []byte) *Blob {
	if len(data) == 0 {
		return &Blob{MIMEType: mimeType}
	}
	return &Blob{MIMEType: mimeType, Data: base64.StdEncoding.EncodeToString(data)}
}

// GenerateContentRequest is the subset of Gemini generateContent request fields
// that affect token accounting. ExtraRequestFields preserves fast-moving API
// fields without forcing this adapter to model every provider schema.
type GenerateContentRequest struct {
	Contents           []Content      `json:"contents,omitempty"`
	SystemInstruction  *Content       `json:"systemInstruction,omitempty"`
	Tools              any            `json:"tools,omitempty"`
	ToolConfig         any            `json:"toolConfig,omitempty"`
	SafetySettings     any            `json:"safetySettings,omitempty"`
	GenerationConfig   any            `json:"generationConfig,omitempty"`
	CachedContent      string         `json:"cachedContent,omitempty"`
	ExtraRequestFields map[string]any `json:"-"`
}

// CountTokensRequest is a provider-side countTokens request.
//
// For the Gemini Developer API, GenerateContentRequest is sent inside the
// generateContentRequest wrapper. For Vertex AI, the same fields are sent at the
// top level because that endpoint accepts the generation fields directly.
type CountTokensRequest struct {
	Model                  string
	Contents               []Content
	GenerateContentRequest *GenerateContentRequest
	ExtraRequestFields     map[string]any
}

// ModalityTokenCount is a provider token count broken down by modality.
type ModalityTokenCount struct {
	Modality   string `json:"modality,omitempty"`
	TokenCount int    `json:"tokenCount,omitempty"`
}

// CountTokensResponse is returned by Gemini/Vertex countTokens.
//
// Multimodal counts returned by countTokens can still be estimates; final usage
// metadata from a generate call is the authoritative post-execution record.
type CountTokensResponse struct {
	TotalTokens             int                  `json:"totalTokens,omitempty"`
	TotalBillableCharacters int                  `json:"totalBillableCharacters,omitempty"`
	CachedContentTokenCount int                  `json:"cachedContentTokenCount,omitempty"`
	PromptTokensDetails     []ModalityTokenCount `json:"promptTokensDetails,omitempty"`
	CacheTokensDetails      []ModalityTokenCount `json:"cacheTokensDetails,omitempty"`
	Raw                     json.RawMessage      `json:"-"`
}

// UsageMetadata is the authoritative token accounting block returned by Gemini
// generation responses after execution.
type UsageMetadata struct {
	PromptTokenCount           int                  `json:"promptTokenCount,omitempty"`
	CachedContentTokenCount    int                  `json:"cachedContentTokenCount,omitempty"`
	CandidatesTokenCount       int                  `json:"candidatesTokenCount,omitempty"`
	ToolUsePromptTokenCount    int                  `json:"toolUsePromptTokenCount,omitempty"`
	ThoughtsTokenCount         int                  `json:"thoughtsTokenCount,omitempty"`
	TotalTokenCount            int                  `json:"totalTokenCount,omitempty"`
	PromptTokensDetails        []ModalityTokenCount `json:"promptTokensDetails,omitempty"`
	CacheTokensDetails         []ModalityTokenCount `json:"cacheTokensDetails,omitempty"`
	CandidatesTokensDetails    []ModalityTokenCount `json:"candidatesTokensDetails,omitempty"`
	ToolUsePromptTokensDetails []ModalityTokenCount `json:"toolUsePromptTokensDetails,omitempty"`
	ServiceTier                string               `json:"serviceTier,omitempty"`
	TrafficType                string               `json:"trafficType,omitempty"`
	Raw                        json.RawMessage      `json:"-"`
}

// CountContentTokens counts structured Gemini contents with the provider API.
func (c Client) CountContentTokens(ctx context.Context, model string, contents []Content) (CountTokensResponse, error) {
	return c.CountTokens(ctx, CountTokensRequest{Model: model, Contents: contents})
}

// CountGenerateContentRequest counts a full generation-shaped request with the
// provider API, including system instructions, tools, cached content, and
// multimodal content blocks supported by the provider endpoint.
func (c Client) CountGenerateContentRequest(ctx context.Context, model string, req GenerateContentRequest) (CountTokensResponse, error) {
	return c.CountTokens(ctx, CountTokensRequest{Model: model, GenerateContentRequest: &req})
}

// CountTokens calls Gemini Developer API or Vertex AI countTokens depending on
// the client configuration. Project or Location selects Vertex AI; otherwise the
// Gemini Developer API is used.
func (c Client) CountTokens(ctx context.Context, req CountTokensRequest) (CountTokensResponse, error) {
	if strings.TrimSpace(req.Model) == "" {
		return CountTokensResponse{}, errors.New("omnitoken gemini: model is required")
	}
	if req.GenerateContentRequest == nil && len(req.Contents) == 0 && len(req.ExtraRequestFields) == 0 {
		return CountTokensResponse{}, errors.New("omnitoken gemini: contents or generate content request is required")
	}

	body, err := c.countRequestBody(req)
	if err != nil {
		return CountTokensResponse{}, err
	}
	endpoint, err := c.countTokensURL(req.Model)
	if err != nil {
		return CountTokensResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return CountTokensResponse{}, err
	}
	httpReq.Header.Set("content-type", "application/json")
	if err := c.authorize(httpReq); err != nil {
		return CountTokensResponse{}, err
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return CountTokensResponse{}, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CountTokensResponse{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return CountTokensResponse{}, fmt.Errorf("omnitoken gemini: countTokens returned %s: %s", resp.Status, string(respBody))
	}
	var result CountTokensResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return CountTokensResponse{}, err
	}
	result.Raw = append(result.Raw[:0], respBody...)
	return result, nil
}

// ParseUsageMetadata extracts usageMetadata from a Gemini generate response. If
// data is already a usageMetadata object, it is parsed directly.
func ParseUsageMetadata(data []byte) (UsageMetadata, error) {
	var wrapper struct {
		UsageMetadata *UsageMetadata `json:"usageMetadata"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return UsageMetadata{}, err
	}
	if wrapper.UsageMetadata != nil {
		wrapper.UsageMetadata.Raw = cloneRaw(data)
		return *wrapper.UsageMetadata, nil
	}

	var usage UsageMetadata
	if err := json.Unmarshal(data, &usage); err != nil {
		return UsageMetadata{}, err
	}
	usage.Raw = cloneRaw(data)
	return usage, nil
}

func (c Client) countRequestBody(req CountTokensRequest) ([]byte, error) {
	if c.useVertex() {
		body := map[string]any{}
		if req.GenerateContentRequest != nil {
			mergeMap(body, generateContentBody(*req.GenerateContentRequest))
		} else {
			body["contents"] = req.Contents
		}
		for key, value := range req.ExtraRequestFields {
			body[key] = value
		}
		return json.Marshal(body)
	}

	body := map[string]any{}
	if req.GenerateContentRequest != nil {
		body["generateContentRequest"] = generateContentBody(*req.GenerateContentRequest)
	} else {
		body["contents"] = req.Contents
	}
	for key, value := range req.ExtraRequestFields {
		body[key] = value
	}
	return json.Marshal(body)
}

func generateContentBody(req GenerateContentRequest) map[string]any {
	body := map[string]any{}
	if len(req.Contents) > 0 {
		body["contents"] = req.Contents
	}
	if req.SystemInstruction != nil {
		body["systemInstruction"] = req.SystemInstruction
	}
	if req.Tools != nil {
		body["tools"] = req.Tools
	}
	if req.ToolConfig != nil {
		body["toolConfig"] = req.ToolConfig
	}
	if req.SafetySettings != nil {
		body["safetySettings"] = req.SafetySettings
	}
	if req.GenerationConfig != nil {
		body["generationConfig"] = req.GenerationConfig
	}
	if req.CachedContent != "" {
		body["cachedContent"] = req.CachedContent
	}
	for key, value := range req.ExtraRequestFields {
		body[key] = value
	}
	return body
}

func (c Client) countTokensURL(model string) (string, error) {
	if c.useVertex() {
		return c.vertexCountTokensURL(model)
	}
	return c.developerCountTokensURL(model)
}

func (c Client) developerCountTokensURL(model string) (string, error) {
	baseURL := strings.TrimRight(c.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultDeveloperBaseURL
	}
	version := c.APIVersion
	if version == "" {
		version = "v1beta"
	}
	resource := model
	if !strings.Contains(resource, "/") {
		resource = "models/" + resource
	}
	return baseURL + "/" + strings.Trim(version, "/") + "/" + escapeResourcePath(resource) + ":countTokens", nil
}

func (c Client) vertexCountTokensURL(model string) (string, error) {
	project := strings.TrimSpace(c.Project)
	location := strings.TrimSpace(c.Location)
	if project == "" {
		return "", errors.New("omnitoken gemini: project is required for Vertex AI")
	}
	if location == "" {
		return "", errors.New("omnitoken gemini: location is required for Vertex AI")
	}
	baseURL := strings.TrimRight(c.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://" + location + "-aiplatform.googleapis.com"
	}
	version := c.APIVersion
	if version == "" {
		version = defaultVertexVersion
	}
	resource := model
	if !strings.HasPrefix(resource, "projects/") {
		resource = "projects/" + project + "/locations/" + location + "/publishers/google/models/" + strings.TrimPrefix(resource, "models/")
	}
	return baseURL + "/" + strings.Trim(version, "/") + "/" + escapeResourcePath(resource) + ":countTokens", nil
}

func (c Client) authorize(req *http.Request) error {
	if c.useVertex() {
		if c.BearerToken == "" {
			return errors.New("omnitoken gemini: bearer token is required for Vertex AI")
		}
		req.Header.Set("authorization", "Bearer "+c.BearerToken)
		return nil
	}
	if c.APIKey != "" {
		req.Header.Set("x-goog-api-key", c.APIKey)
		return nil
	}
	if c.BearerToken != "" {
		req.Header.Set("authorization", "Bearer "+c.BearerToken)
		return nil
	}
	return errors.New("omnitoken gemini: API key or bearer token is required")
}

func (c Client) useVertex() bool {
	return c.Project != "" || c.Location != ""
}

func mergeMap(dst, src map[string]any) {
	for key, value := range src {
		dst[key] = value
	}
}

func escapeResourcePath(resource string) string {
	parts := strings.Split(strings.Trim(resource, "/"), "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func cloneRaw(data []byte) json.RawMessage {
	if len(data) == 0 {
		return nil
	}
	return append(json.RawMessage(nil), data...)
}
