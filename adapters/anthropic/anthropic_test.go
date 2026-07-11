package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCountMessageTokens(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages/count_tokens" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Fatalf("missing API key header")
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Fatalf("missing version header")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["system"] != "be concise" {
			t.Fatalf("system = %#v", body["system"])
		}
		if _, ok := body["tools"].([]any); !ok {
			t.Fatalf("tools = %#v", body["tools"])
		}
		if body["service_tier"] != "auto" {
			t.Fatalf("extra field = %#v", body["service_tier"])
		}
		_, _ = w.Write([]byte(`{"input_tokens":42}`))
	}))
	defer server.Close()

	client := Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()}
	result, err := client.CountMessageTokens(context.Background(), CountRequest{
		Model:    "claude-sonnet-test",
		System:   "be concise",
		Messages: []Message{{Role: "user", Content: "hello"}},
		Tools:    []map[string]any{{"name": "lookup"}},
		ExtraRequestFields: map[string]any{
			"service_tier": "auto",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InputTokens != 42 {
		t.Fatalf("InputTokens = %d", result.InputTokens)
	}
}

func TestParseUsage(t *testing.T) {
	usage, err := ParseUsage([]byte(`{"usage":{"input_tokens":10,"output_tokens":4,"cache_creation_input_tokens":2,"cache_read_input_tokens":3,"server_tool_use":{"web_search_requests":1}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if usage.InputTokens != 10 || usage.OutputTokens != 4 || usage.CacheCreationInputTokens != 2 || usage.CacheReadInputTokens != 3 {
		t.Fatalf("usage = %+v", usage)
	}
	if usage.ServerToolUse["web_search_requests"] != 1 {
		t.Fatalf("ServerToolUse = %+v", usage.ServerToolUse)
	}

	direct, err := ParseUsage([]byte(`{"input_tokens":1,"output_tokens":2}`))
	if err != nil {
		t.Fatal(err)
	}
	if direct.InputTokens != 1 || direct.OutputTokens != 2 {
		t.Fatalf("direct usage = %+v", direct)
	}
}

func TestCountMessageTokensValidation(t *testing.T) {
	client := Client{}
	_, err := client.CountMessageTokens(context.Background(), CountRequest{})
	if err == nil || !strings.Contains(err.Error(), "API key") {
		t.Fatalf("missing API key err = %v", err)
	}
	client.APIKey = "test"
	_, err = client.CountMessageTokens(context.Background(), CountRequest{})
	if err == nil || !strings.Contains(err.Error(), "model") {
		t.Fatalf("missing model err = %v", err)
	}
}

func TestCountMessageTokensHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusBadRequest)
	}))
	defer server.Close()

	client := Client{APIKey: "test", BaseURL: server.URL, HTTPClient: server.Client()}
	_, err := client.CountMessageTokens(context.Background(), CountRequest{
		Model:    "claude-sonnet-test",
		Messages: []Message{{Role: "user", Content: "hello"}},
	})
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Fatalf("HTTP error = %v", err)
	}
}
