package service

import (
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/steamclient"
	steam "github.com/gofurry/steam-go"
)

func TestSingleGameTaskTypesExcludePlayers(t *testing.T) {
	tasks := singleGameTaskTypes()
	if len(tasks) != 2 || tasks[0] != domain.TaskDetails || tasks[1] != domain.TaskNews {
		t.Fatalf("single-game tasks = %v, want [details news]", tasks)
	}
	for _, task := range tasks {
		if task == domain.TaskPlayers {
			t.Fatal("single-game collection must not include players")
		}
	}
}

func TestSteamRequestDiagnosticFieldsAreSanitized(t *testing.T) {
	event := steamclient.Event{
		Bucket: steamclient.BucketOfficialAPI, TrafficClass: steam.TrafficClassOfficialAPI,
		Method: "GET", Host: "user:secret@api.steampowered.com/path?key=secret",
		Path: "/IStoreBrowseService/GetItems/v1?key=secret", StatusCode: 429,
		ErrorKind: "http_status", Attempts: 2, BlockDetected: true,
		Duration: 50 * time.Millisecond, CooldownUntil: time.Date(2026, 9, 4, 1, 2, 3, 0, time.UTC),
	}
	if !steamRequestNeedsDiagnostic(event) {
		t.Fatal("degraded request was not selected for diagnostics")
	}
	fields := steamRequestDiagnosticFields(event)
	if fields["host"] != "api.steampowered.com" || fields["path"] != "/IStoreBrowseService/GetItems/v1" {
		t.Fatalf("observer fields were not sanitized: %#v", fields)
	}
	if strings.Contains(fields["host"].(string)+fields["path"].(string), "secret") {
		t.Fatalf("observer fields leaked secrets: %#v", fields)
	}
	for _, key := range []string{"bucket", "traffic_class", "method", "host", "path", "status_code", "error_kind", "attempts", "block_detected", "duration", "cooldown_until"} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("diagnostic field %q missing: %#v", key, fields)
		}
	}
}

func TestSteamRequestDiagnosticIgnoresHealthyRequests(t *testing.T) {
	if steamRequestNeedsDiagnostic(steamclient.Event{StatusCode: 200}) {
		t.Fatal("healthy request should not create diagnostic log noise")
	}
}
