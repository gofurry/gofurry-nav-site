package service

import (
	"sync"
	"testing"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/roof/env"
)

func TestDevelopmentHomeCacheRequiresExplicitDebugOptIn(t *testing.T) {
	cfg := env.GetServerConfig()
	previous := cfg.Server
	t.Cleanup(func() { cfg.Server = previous })
	for _, tc := range []struct {
		mode    string
		seconds int
		enabled bool
	}{
		{"debug", 0, false},
		{"release", 0, false},
		{"release", 10, false},
		{"production", 10, false},
		{"debug", 10, true},
	} {
		cfg.Server.Mode = tc.mode
		cfg.Server.DevelopmentHomeCacheSeconds = tc.seconds
		svc := NewReadModelServiceWithReader(nil)
		if (svc.developmentHome != nil) != tc.enabled {
			t.Fatalf("mode=%s seconds=%d: unexpected development cache state", tc.mode, tc.seconds)
		}
	}
}

func TestDevelopmentHomeCacheExpiryIsolationAndFailure(t *testing.T) {
	now := time.Now()
	cache := developmentHomeCache{ttl: 10 * time.Second, now: func() time.Time { return now }}
	calls := 0
	available := true
	fetch := func() (v2models.GameV2HomeReadModel, bool) {
		calls++
		return v2models.GameV2HomeReadModel{Panel: v2models.GameV2PanelReadModel{LatestGames: []v2models.GameV2ListItem{{ID: "1"}}}}, available
	}
	first, _ := cache.load("zh", fetch)
	first.Panel.LatestGames[0].ID = "mutated"
	second, hit := cache.load("zh", fetch)
	if !hit || calls != 1 || second.Panel.LatestGames[0].ID != "1" {
		t.Fatal("cached response aliases caller data or repeated remote read")
	}
	cache.load("en", fetch)
	if calls != 2 {
		t.Fatal("language cache collision")
	}
	now = now.Add(10 * time.Second)
	available = false
	if _, hit := cache.load("zh", fetch); hit {
		t.Fatal("expired cache masked upstream failure")
	}
	if _, hit := cache.load("zh", fetch); hit || calls != 4 {
		t.Fatal("failure was cached")
	}
	available = true
	cache.load("zh", fetch)
	cache.invalidate("zh")
	cache.load("zh", fetch)
	if calls != 6 {
		t.Fatal("explicit refresh did not invalidate local entry")
	}
}

func TestDevelopmentHomeCacheCoalescesConcurrentReads(t *testing.T) {
	cache := developmentHomeCache{ttl: time.Second, now: time.Now}
	calls := 0
	var workers sync.WaitGroup
	for range 8 {
		workers.Go(func() {
			cache.load("zh", func() (v2models.GameV2HomeReadModel, bool) { calls++; return v2models.GameV2HomeReadModel{}, true })
		})
	}
	workers.Wait()
	if calls != 1 {
		t.Fatalf("remote reads=%d", calls)
	}
}
