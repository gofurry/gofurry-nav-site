package observation

import (
	"strings"
	"testing"

	"github.com/gofurry/gofurry-nav-collector/roof/env"
)

func TestLatestKey(t *testing.T) {
	got := LatestKey(ProtocolPing, 123)
	want := "collector:v2:latest:ping:123"
	if got != want {
		t.Fatalf("LatestKey() = %q, want %q", got, want)
	}
}

func TestMarshalPayloadNormalizesNilPayload(t *testing.T) {
	payload, err := marshalPayload(nil)
	if err != nil {
		t.Fatalf("marshalPayload() error = %v", err)
	}
	if string(payload) != "{}" {
		t.Fatalf("marshalPayload(nil) = %q, want {}", string(payload))
	}
}

func TestObservationPayloadSizeLimit(t *testing.T) {
	payload, err := marshalPayload(map[string]string{
		"value": string(make([]byte, maxObservationPayloadBytes)),
	})
	if err != nil {
		t.Fatalf("marshalPayload() error = %v", err)
	}
	if len(payload) <= maxObservationPayloadBytes {
		t.Fatalf("test payload should exceed limit, got %d <= %d", len(payload), maxObservationPayloadBytes)
	}
}

func TestEnrichPayloadAddsRunIdentityToObjectPayload(t *testing.T) {
	payload := map[string]any{"status": "ok"}
	got := enrichPayload(payload, "collector-a", "ping-1").(map[string]any)
	if got["collector_id"] != "collector-a" || got["job_id"] != "ping-1" || got["status"] != "ok" {
		t.Fatalf("enriched payload wrong: %+v", got)
	}
	if _, exists := payload["collector_id"]; exists {
		t.Fatal("enrichPayload should not mutate original map")
	}
}

func TestEnrichPayloadHandlesNilAndNonObjectPayloadSafely(t *testing.T) {
	got := enrichPayload(nil, "collector-a", "ping-1").(map[string]any)
	if got["collector_id"] != "collector-a" || got["job_id"] != "ping-1" {
		t.Fatalf("nil payload enrichment wrong: %+v", got)
	}
	if got := enrichPayload("raw", "collector-a", "ping-1"); got != "raw" {
		t.Fatalf("non-object payload should stay unchanged, got %+v", got)
	}
}

func TestSaveIfEnabledRejectsOversizedPayloadBeforeWriting(t *testing.T) {
	oldV2 := env.GetServerConfig().Collector.V2
	env.GetServerConfig().Collector.V2 = env.CollectorV2Config{
		Enabled: true,
		Protocols: env.CollectorProtocols{
			Ping: true,
		},
	}
	t.Cleanup(func() {
		env.GetServerConfig().Collector.V2 = oldV2
	})

	err := SaveIfEnabled(Input{
		SiteID:   1,
		Target:   "example.com",
		Protocol: ProtocolPing,
		Payload: map[string]string{
			"value": strings.Repeat("x", maxObservationPayloadBytes),
		},
	})
	if err == nil {
		t.Fatal("SaveIfEnabled() should reject oversized payload")
	}
}
