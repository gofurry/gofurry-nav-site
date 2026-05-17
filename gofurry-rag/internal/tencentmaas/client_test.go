package tencentmaas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientComplete(t *testing.T) {
	var got requestPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer secret" {
			t.Fatalf("authorization = %q", auth)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{
			"id":"1",
			"object":"chat.completion",
			"model":"deepseek-v4-flash",
			"choices":[{"index":0,"message":{"role":"assistant","content":"hello","reasoning_content":"think"},"finish_reason":"stop"}],
			"usage":{"prompt_tokens":11,"completion_tokens":22,"total_tokens":33,"prompt_tokens_details":{"cached_tokens":4},"completion_tokens_details":{"reasoning_tokens":5}}
		}`)
	}))
	defer server.Close()

	client := New(server.URL+"/v1", "secret", "deepseek-v4-flash", time.Second, 0.2, 0.8, 1024, "low")
	result, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "你好"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Model != "deepseek-v4-flash" || got.Stream {
		t.Fatalf("payload = %+v", got)
	}
	if got.Temperature != 0.2 || got.TopP != 0.8 || got.MaxTokens != 1024 || got.ReasoningEffort != "low" {
		t.Fatalf("payload = %+v", got)
	}
	if result.Answer != "hello" || result.Reasoning != "think" || result.PromptTokens != 11 || result.CachedTokens != 4 {
		t.Fatalf("result = %+v", result)
	}
}

func TestClientStream(t *testing.T) {
	var got requestPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer is not flushable")
		}
		_, _ = fmt.Fprint(w, "data: {\"id\":\"1\",\"model\":\"deepseek-v4-flash\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"content\":\"hello\"},\"finish_reason\":null}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":2,\"total_tokens\":3,\"prompt_tokens_details\":{\"cached_tokens\":0},\"completion_tokens_details\":{\"reasoning_tokens\":0}}}\n\n")
		flusher.Flush()
		_, _ = fmt.Fprint(w, "data: {\"id\":\"2\",\"model\":\"deepseek-v4-flash\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\" world\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":2,\"total_tokens\":3,\"prompt_tokens_details\":{\"cached_tokens\":0},\"completion_tokens_details\":{\"reasoning_tokens\":0}}}\n\n")
		flusher.Flush()
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	client := New(server.URL+"/v1", "secret", "deepseek-v4-flash", time.Second, 0.2, 0.8, 1024, "low")
	var deltas []string
	result, err := client.Stream(context.Background(), []Message{{Role: "user", Content: "你好"}}, func(text string) error {
		deltas = append(deltas, text)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Stream != true || got.Model != "deepseek-v4-flash" {
		t.Fatalf("payload = %+v", got)
	}
	if strings.Join(deltas, "") != "hello world" {
		t.Fatalf("deltas = %#v", deltas)
	}
	if result.Answer != "hello world" || result.Model != "deepseek-v4-flash" {
		t.Fatalf("result = %+v", result)
	}
}

func TestClientErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":{"message":"bad request"}}`)
	}))
	defer server.Close()

	client := New(server.URL+"/v1", "secret", "deepseek-v4-flash", time.Second, 0.2, 0.8, 1024, "low")
	_, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "你好"}})
	if err == nil || !strings.Contains(err.Error(), "bad request") {
		t.Fatalf("err = %v", err)
	}
}

func TestClientErrorResponseIsCapped(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, strings.Repeat("x", maxErrorBodyBytes+1024))
	}))
	defer server.Close()

	client := New(server.URL+"/v1", "secret", "deepseek-v4-flash", time.Second, 0.2, 0.8, 1024, "low")
	_, err := client.Complete(context.Background(), []Message{{Role: "user", Content: "你好"}})
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Count(err.Error(), "x") > maxErrorBodyBytes {
		t.Fatalf("error body was not capped: %d", strings.Count(err.Error(), "x"))
	}
}

func TestClientHealthProbe(t *testing.T) {
	var got requestPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer secret" {
			t.Fatalf("authorization = %q", auth)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"id":"1","object":"chat.completion","model":"deepseek-v4-flash","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2,"prompt_tokens_details":{"cached_tokens":0},"completion_tokens_details":{"reasoning_tokens":0}}}`)
	}))
	defer server.Close()

	client := New(server.URL+"/v1", "secret", "deepseek-v4-flash", time.Second, 0.2, 0.8, 1024, "low")
	if err := client.Health(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got.Stream {
		t.Fatalf("health probe should not stream: %+v", got)
	}
	if got.MaxTokens != 1 {
		t.Fatalf("health probe max_tokens = %d", got.MaxTokens)
	}
	if len(got.Messages) != 2 || got.Messages[1].Content != "ping" {
		t.Fatalf("health probe messages = %+v", got.Messages)
	}
}
