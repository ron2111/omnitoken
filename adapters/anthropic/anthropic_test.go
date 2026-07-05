package anthropic

import (
	"context"
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
		_, _ = w.Write([]byte(`{"input_tokens":42}`))
	}))
	defer server.Close()

	client := Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()}
	result, err := client.CountMessageTokens(context.Background(), CountRequest{
		Model:    "claude-sonnet-test",
		Messages: []Message{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InputTokens != 42 {
		t.Fatalf("InputTokens = %d", result.InputTokens)
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
