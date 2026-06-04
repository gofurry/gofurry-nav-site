package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/service"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/search/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/gofurry/gofurry-nav-backend/common/util"
	"golang.org/x/sync/singleflight"
)

const (
	searchSuggestionMaxQueryLen = 128
	searchSuggestionCacheTTL    = 90 * time.Second
)

type suggestionProvider interface {
	GetBaiduSuggestion(q string) ([]string, common.GFError)
	GetBingSuggestion(q string) ([]string, common.GFError)
	GetGoogleSuggestion(q string) ([]string, common.GFError)
	GetBiliBiliSuggestion(q string) ([]string, common.GFError)
}

type suggestionCache interface {
	Get(key string) ([]string, bool)
	Set(key string, suggestions []string)
}

type redisSuggestionCache struct{}

func (redisSuggestionCache) Get(key string) ([]string, bool) {
	raw, err := cs.GetString(key)
	if err != nil || raw == "" {
		return nil, false
	}
	var suggestions []string
	if jsonErr := sonic.Unmarshal([]byte(raw), &suggestions); jsonErr != nil {
		log.Warn("search suggestion cache unmarshal error:", jsonErr)
		return nil, false
	}
	return copySuggestions(suggestions), true
}

func (redisSuggestionCache) Set(key string, suggestions []string) {
	payload, err := sonic.Marshal(suggestions)
	if err != nil {
		return
	}
	_ = cs.SetExpire(key, string(payload), searchSuggestionCacheTTL)
}

type searchService struct {
	provider suggestionProvider
	cache    suggestionCache
	group    singleflight.Group
	now      func() time.Time
}

var (
	searchSingleton = &searchService{}
	searchMu        sync.Mutex
)

func GetSearchService() *searchService {
	searchMu.Lock()
	defer searchMu.Unlock()
	if searchSingleton.provider == nil {
		searchSingleton.provider = service.GetNavPageService()
	}
	if searchSingleton.cache == nil {
		searchSingleton.cache = redisSuggestionCache{}
	}
	if searchSingleton.now == nil {
		searchSingleton.now = time.Now
	}
	return searchSingleton
}

func newSearchService(provider suggestionProvider, cache suggestionCache, now func() time.Time) *searchService {
	return &searchService{provider: provider, cache: cache, now: now}
}

func (svc *searchService) GetSearchSuggestions(engine string, query string) models.SearchSuggestionsResponse {
	engine = normalizeSuggestionEngine(engine)
	query = normalizeSuggestionQuery(query)
	response := models.SearchSuggestionsResponse{
		SchemaVersion: models.SearchSuggestionsSchemaVersion,
		GeneratedAt:   svc.clock()(),
		State:         models.SearchSuggestionsStateEmpty,
		Engine:        engine,
		Query:         query,
		Suggestions:   []string{},
		CacheState:    models.SearchSuggestionsCacheMiss,
	}

	if engine == "" {
		response.State = models.SearchSuggestionsStateError
		response.ReasonMessages = []string{"unsupported search engine"}
		return response
	}
	if query == "" {
		return response
	}

	cacheKey := searchSuggestionCacheKey(engine, query)
	if cached, ok := svc.cacheStore().Get(cacheKey); ok {
		response.CacheState = models.SearchSuggestionsCacheHit
		response.Suggestions = cached
		if len(cached) > 0 {
			response.State = models.SearchSuggestionsStateReady
		}
		return response
	}

	result, err, _ := svc.group.Do(cacheKey, func() (any, error) {
		if cached, ok := svc.cacheStore().Get(cacheKey); ok {
			return cachedSearchSuggestions{items: cached, hit: true}, nil
		}
		items, fetchErr := svc.fetchSuggestions(engine, query)
		if fetchErr != nil {
			return cachedSearchSuggestions{items: []string{}, hit: false}, errors.New(fetchErr.GetMsg())
		}
		items = sanitizeSuggestions(items)
		svc.cacheStore().Set(cacheKey, items)
		return cachedSearchSuggestions{items: items, hit: false}, nil
	})
	if err != nil {
		response.State = models.SearchSuggestionsStateError
		response.ReasonMessages = []string{err.Error()}
		return response
	}

	data := result.(cachedSearchSuggestions)
	response.Suggestions = copySuggestions(data.items)
	if data.hit {
		response.CacheState = models.SearchSuggestionsCacheHit
	}
	if len(response.Suggestions) > 0 {
		response.State = models.SearchSuggestionsStateReady
	}
	return response
}

func (svc *searchService) fetchSuggestions(engine string, query string) ([]string, common.GFError) {
	switch engine {
	case "baidu":
		return svc.source().GetBaiduSuggestion(query)
	case "bing":
		return svc.source().GetBingSuggestion(query)
	case "google":
		return svc.source().GetGoogleSuggestion(query)
	case "bilibili":
		return svc.source().GetBiliBiliSuggestion(query)
	default:
		return []string{}, nil
	}
}

func (svc *searchService) source() suggestionProvider {
	if svc != nil && svc.provider != nil {
		return svc.provider
	}
	return service.GetNavPageService()
}

func (svc *searchService) cacheStore() suggestionCache {
	if svc != nil && svc.cache != nil {
		return svc.cache
	}
	return redisSuggestionCache{}
}

func (svc *searchService) clock() func() time.Time {
	if svc != nil && svc.now != nil {
		return svc.now
	}
	return time.Now
}

type cachedSearchSuggestions struct {
	items []string
	hit   bool
}

func normalizeSuggestionEngine(engine string) string {
	switch strings.ToLower(strings.TrimSpace(engine)) {
	case "baidu", "bing", "google", "bilibili":
		return strings.ToLower(strings.TrimSpace(engine))
	default:
		return ""
	}
}

func normalizeSuggestionQuery(q string) string {
	q = strings.TrimSpace(q)
	runes := []rune(q)
	if len(runes) <= searchSuggestionMaxQueryLen {
		return q
	}
	return string(runes[:searchSuggestionMaxQueryLen])
}

func searchSuggestionCacheKey(engine string, query string) string {
	return "nav:v2:search:suggestions:" + engine + ":" + util.CreateMD5(query)
}

func sanitizeSuggestions(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func copySuggestions(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	copied := make([]string, len(items))
	copy(copied, items)
	return copied
}

type redisSuggestionRateLimiter struct {
	now func() time.Time
}

func NewRedisSuggestionRateLimiter() *redisSuggestionRateLimiter {
	return &redisSuggestionRateLimiter{now: time.Now}
}

func (limiter *redisSuggestionRateLimiter) Allow(ip string) (bool, int64) {
	client := cs.GetRedisService()
	if client == nil {
		return true, 0
	}
	ip = strings.TrimSpace(ip)
	if ip == "" {
		ip = "unknown"
	}
	now := time.Now
	if limiter != nil && limiter.now != nil {
		now = limiter.now
	}
	window := now().Unix() / int64(SearchSuggestionRateWindow/time.Second)
	key := "nav:v2:search:suggestions:rate:" + util.CreateMD5(ip) + ":" + fmt.Sprint(window)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		log.Warn("search suggestion rate limiter incr error:", err)
		return true, 0
	}
	if count == 1 {
		_ = client.Expire(ctx, key, SearchSuggestionRateWindow+time.Minute).Err()
	}
	if count > SearchSuggestionRateLimit {
		return false, int64(SearchSuggestionRateWindow / time.Second)
	}
	return true, 0
}

const (
	SearchSuggestionRateLimit  = int64(30)
	SearchSuggestionRateWindow = 30 * time.Minute
)
