package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientCountContentTokensDeveloper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta/models/gemini-2.5-flash:countTokens" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("x-goog-api-key") != "test-key" {
			t.Fatalf("missing API key header")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		contents, ok := body["contents"].([]any)
		if !ok || len(contents) != 1 {
			t.Fatalf("contents = %#v", body["contents"])
		}
		parts := contents[0].(map[string]any)["parts"].([]any)
		if got := parts[0].(map[string]any)["text"]; got != "hello" {
			t.Fatalf("text part = %#v", got)
		}
		inlineData := parts[1].(map[string]any)["inlineData"].(map[string]any)
		if inlineData["mimeType"] != "image/png" || inlineData["data"] != "AQI=" {
			t.Fatalf("inlineData = %#v", inlineData)
		}
		fileData := parts[2].(map[string]any)["fileData"].(map[string]any)
		if fileData["fileUri"] != "gs://bucket/video.mp4" {
			t.Fatalf("fileData = %#v", fileData)
		}
		_, _ = w.Write([]byte(`{"totalTokens":42,"cachedContentTokenCount":7,"promptTokensDetails":[{"modality":"TEXT","tokenCount":12},{"modality":"IMAGE","tokenCount":30}],"cacheTokensDetails":[{"modality":"TEXT","tokenCount":7}]}`))
	}))
	defer server.Close()

	client := Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()}
	result, err := client.CountContentTokens(context.Background(), "gemini-2.5-flash", []Content{{
		Role: "user",
		Parts: []Part{
			{Text: "hello"},
			{InlineData: InlineDataFromBytes("image/png", []byte{1, 2})},
			{FileData: &FileData{MIMEType: "video/mp4", FileURI: "gs://bucket/video.mp4"}},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalTokens != 42 || result.CachedContentTokenCount != 7 {
		t.Fatalf("result = %+v", result)
	}
	if len(result.PromptTokensDetails) != 2 || result.PromptTokensDetails[1].Modality != "IMAGE" {
		t.Fatalf("PromptTokensDetails = %+v", result.PromptTokensDetails)
	}
}

func TestClientCountGenerateContentRequestDeveloper(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		wrapped, ok := body["generateContentRequest"].(map[string]any)
		if !ok {
			t.Fatalf("missing generateContentRequest: %#v", body)
		}
		if wrapped["cachedContent"] != "cachedContents/test" {
			t.Fatalf("cachedContent = %#v", wrapped["cachedContent"])
		}
		if _, ok := wrapped["systemInstruction"].(map[string]any); !ok {
			t.Fatalf("systemInstruction = %#v", wrapped["systemInstruction"])
		}
		if _, ok := wrapped["tools"].([]any); !ok {
			t.Fatalf("tools = %#v", wrapped["tools"])
		}
		if wrapped["responseMimeType"] != "application/json" {
			t.Fatalf("extra field = %#v", wrapped["responseMimeType"])
		}
		_, _ = w.Write([]byte(`{"totalTokens":11}`))
	}))
	defer server.Close()

	client := Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()}
	result, err := client.CountGenerateContentRequest(context.Background(), "gemini-2.5-flash", GenerateContentRequest{
		Contents:          []Content{{Role: "user", Parts: []Part{{Text: "hello"}}}},
		SystemInstruction: &Content{Parts: []Part{{Text: "be concise"}}},
		Tools:             []map[string]any{{"functionDeclarations": []any{}}},
		CachedContent:     "cachedContents/test",
		ExtraRequestFields: map[string]any{
			"responseMimeType": "application/json",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalTokens != 11 {
		t.Fatalf("TotalTokens = %d", result.TotalTokens)
	}
}

func TestClientCountTokensVertex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/v1/projects/project-1/locations/us-central1/publishers/google/models/gemini-2.5-flash:countTokens"
		if r.URL.Path != wantPath {
			t.Fatalf("path = %s, want %s", r.URL.Path, wantPath)
		}
		if r.Header.Get("authorization") != "Bearer test-token" {
			t.Fatalf("authorization = %q", r.Header.Get("authorization"))
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if _, ok := body["generateContentRequest"]; ok {
			t.Fatalf("Vertex body should not use wrapper: %#v", body)
		}
		if _, ok := body["systemInstruction"].(map[string]any); !ok {
			t.Fatalf("systemInstruction = %#v", body["systemInstruction"])
		}
		_, _ = w.Write([]byte(`{"totalTokens":9,"totalBillableCharacters":96}`))
	}))
	defer server.Close()

	client := Client{BearerToken: "test-token", BaseURL: server.URL, Project: "project-1", Location: "us-central1", HTTPClient: server.Client()}
	result, err := client.CountGenerateContentRequest(context.Background(), "gemini-2.5-flash", GenerateContentRequest{
		Contents:          []Content{{Role: "user", Parts: []Part{{Text: "hello"}}}},
		SystemInstruction: &Content{Parts: []Part{{Text: "be concise"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalTokens != 9 || result.TotalBillableCharacters != 96 {
		t.Fatalf("result = %+v", result)
	}
}

func TestClientCountTokensValidation(t *testing.T) {
	client := Client{}
	_, err := client.CountContentTokens(context.Background(), "", []Content{{Parts: []Part{{Text: "hello"}}}})
	if err == nil || !strings.Contains(err.Error(), "model") {
		t.Fatalf("missing model err = %v", err)
	}
	_, err = client.CountContentTokens(context.Background(), "gemini-2.5-flash", nil)
	if err == nil || !strings.Contains(err.Error(), "contents") {
		t.Fatalf("missing contents err = %v", err)
	}
	_, err = client.CountContentTokens(context.Background(), "gemini-2.5-flash", []Content{{Parts: []Part{{Text: "hello"}}}})
	if err == nil || !strings.Contains(err.Error(), "API key") {
		t.Fatalf("missing API key err = %v", err)
	}

	vertex := Client{Project: "p", BearerToken: "token"}
	_, err = vertex.CountContentTokens(context.Background(), "gemini-2.5-flash", []Content{{Parts: []Part{{Text: "hello"}}}})
	if err == nil || !strings.Contains(err.Error(), "location") {
		t.Fatalf("missing location err = %v", err)
	}

	vertex = Client{Project: "p", Location: "us-central1", APIKey: "not-valid-for-vertex"}
	_, err = vertex.CountContentTokens(context.Background(), "gemini-2.5-flash", []Content{{Parts: []Part{{Text: "hello"}}}})
	if err == nil || !strings.Contains(err.Error(), "bearer token") {
		t.Fatalf("missing bearer token err = %v", err)
	}
}

func TestClientCountTokensHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusBadRequest)
	}))
	defer server.Close()

	client := Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()}
	_, err := client.CountContentTokens(context.Background(), "gemini-2.5-flash", []Content{{Parts: []Part{{Text: "hello"}}}})
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Fatalf("HTTP error = %v", err)
	}
}

func TestParseUsageMetadata(t *testing.T) {
	usage, err := ParseUsageMetadata([]byte(`{"usageMetadata":{"promptTokenCount":10,"cachedContentTokenCount":3,"candidatesTokenCount":4,"thoughtsTokenCount":2,"totalTokenCount":16,"promptTokensDetails":[{"modality":"TEXT","tokenCount":10}],"serviceTier":"standard"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if usage.PromptTokenCount != 10 || usage.CachedContentTokenCount != 3 || usage.TotalTokenCount != 16 {
		t.Fatalf("usage = %+v", usage)
	}
	if len(usage.PromptTokensDetails) != 1 || usage.PromptTokensDetails[0].Modality != "TEXT" {
		t.Fatalf("PromptTokensDetails = %+v", usage.PromptTokensDetails)
	}

	direct, err := ParseUsageMetadata([]byte(`{"promptTokenCount":1,"totalTokenCount":2,"trafficType":"ON_DEMAND"}`))
	if err != nil {
		t.Fatal(err)
	}
	if direct.PromptTokenCount != 1 || direct.TrafficType != "ON_DEMAND" {
		t.Fatalf("direct usage = %+v", direct)
	}
}
