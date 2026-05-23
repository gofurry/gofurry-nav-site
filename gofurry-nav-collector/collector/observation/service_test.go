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
