package service

import (
	"sync"
	"time"

	"github.com/bytedance/sonic"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
)

type developmentHomeEntry struct {
	json    []byte
	expires time.Time
}

// Explicit debug-only optimization for a remote development Redis. Only the
// two homepage languages in CN are retained, with no stale-on-error extension.
type developmentHomeCache struct {
	mu      sync.Mutex
	ttl     time.Duration
	now     func() time.Time
	entries map[string]developmentHomeEntry
}

func (svc *ReadModelService) loadHomeCache(key string) (v2models.GameV2HomeReadModel, bool) {
	if svc.developmentHome == nil || (key != "game:v2:home:zh:CN" && key != "game:v2:home:en:CN") {
		return loadGameHomeCache(key)
	}
	return svc.developmentHome.load(key, func() (v2models.GameV2HomeReadModel, bool) { return loadGameHomeCache(key) })
}

func (cache *developmentHomeCache) load(key string, fetch func() (v2models.GameV2HomeReadModel, bool)) (v2models.GameV2HomeReadModel, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if entry, ok := cache.entries[key]; ok && cache.now().Before(entry.expires) {
		var result v2models.GameV2HomeReadModel
		if sonic.Unmarshal(entry.json, &result) == nil {
			return result, true
		}
	}
	delete(cache.entries, key)
	result, hit := fetch()
	if hit {
		if data, err := sonic.Marshal(result); err == nil {
			if cache.entries == nil {
				cache.entries = make(map[string]developmentHomeEntry, 2)
			}
			cache.entries[key] = developmentHomeEntry{json: data, expires: cache.now().Add(cache.ttl)}
		}
	}
	return result, hit
}

func (cache *developmentHomeCache) invalidate(key string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	delete(cache.entries, key)
}
