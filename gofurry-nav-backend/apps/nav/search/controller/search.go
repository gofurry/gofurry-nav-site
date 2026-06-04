package controller

import (
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/search/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/search/service"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/util"
)

type searchApi struct{}

var SearchApi *searchApi

func init() {
	SearchApi = &searchApi{}
}

type suggestionsReader interface {
	GetSearchSuggestions(engine string, query string) models.SearchSuggestionsResponse
}

type suggestionsLimiter interface {
	Allow(ip string) (bool, int64)
}

var (
	searchReaderMu       sync.RWMutex
	searchReaderForTest  suggestionsReader
	searchLimiterMu      sync.RWMutex
	searchLimiterForTest suggestionsLimiter
)

func (api searchApi) GetSearchSuggestions(c fiber.Ctx) error {
	allowed, retryAfter := currentSuggestionsLimiter().Allow(util.GetClientIP(c))
	if !allowed {
		if retryAfter > 0 {
			c.Set("Retry-After", strconv.FormatInt(retryAfter, 10))
		}
		return common.NewResponse(c).ErrorWithCode("搜索建议请求过于频繁，请稍后再试", fiber.StatusTooManyRequests)
	}

	data := currentSuggestionsReader().GetSearchSuggestions(c.Query("engine"), c.Query("q"))
	return common.NewResponse(c).SuccessWithData(data)
}

func currentSuggestionsReader() suggestionsReader {
	searchReaderMu.RLock()
	reader := searchReaderForTest
	searchReaderMu.RUnlock()
	if reader != nil {
		return reader
	}
	return service.GetSearchService()
}

func currentSuggestionsLimiter() suggestionsLimiter {
	searchLimiterMu.RLock()
	limiter := searchLimiterForTest
	searchLimiterMu.RUnlock()
	if limiter != nil {
		return limiter
	}
	return service.NewRedisSuggestionRateLimiter()
}

func setSuggestionsReaderForTest(reader suggestionsReader) func() {
	searchReaderMu.Lock()
	previous := searchReaderForTest
	searchReaderForTest = reader
	searchReaderMu.Unlock()
	return func() {
		searchReaderMu.Lock()
		searchReaderForTest = previous
		searchReaderMu.Unlock()
	}
}

func setSuggestionsLimiterForTest(limiter suggestionsLimiter) func() {
	searchLimiterMu.Lock()
	previous := searchLimiterForTest
	searchLimiterForTest = limiter
	searchLimiterMu.Unlock()
	return func() {
		searchLimiterMu.Lock()
		searchLimiterForTest = previous
		searchLimiterMu.Unlock()
	}
}
