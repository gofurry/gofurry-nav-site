package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

const (
	CollectionCacheTTL     = time.Hour
	collectionRedisTimeout = 200 * time.Millisecond
)

var ErrCollectionQuery = errors.New("invalid collection discovery criteria")

var ErrCollectionNotFound = errors.New("collection not found")
var collectionCodePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type collectionReader interface {
	CountPublishedCollections(context.Context, v2models.CollectionQuery) (int64, error)
	ListPublishedCollections(context.Context, v2models.CollectionQuery) ([]v2models.CollectionRecord, error)
	GetPublishedCollection(context.Context, string) (*v2models.CollectionRecord, error)
	ListPublishedCollectionHomeSlots(context.Context, string) ([]v2models.CollectionRecord, error)
	LoadCollectionProjectionGames(context.Context, []int64, string) (v2models.CollectionGames, error)
	LoadCollectionTimelineDecorations(context.Context, []int64, string) (map[int64]v2models.CollectionTimelineDecoration, error)
}

type collectionCache interface {
	Get(context.Context, string) *redis.StringCmd
	Set(context.Context, string, interface{}, time.Duration) *redis.StatusCmd
}

// CollectionService owns its cache namespace and singleflight group. It neither
// participates in /home's long-lived cache nor depends on Admin invalidation.
type CollectionService struct {
	reader  collectionReader
	cache   collectionCache
	flights singleflight.Group
	Now     func() time.Time
}

func NewCollectionService(reader collectionReader, cache collectionCache) *CollectionService {
	if client, ok := cache.(*redis.Client); ok {
		if client == nil {
			cache = nil
		} else {
			// The runtime client doesn't enable context socket deadlines. Use a
			// timeout clone sharing its pool; do not mutate or close the owner.
			cache = client.WithTimeout(collectionRedisTimeout)
		}
	}
	return &CollectionService{reader: reader, cache: cache, Now: time.Now}
}

func NormalizeCollectionQuery(query v2models.CollectionQuery) v2models.CollectionQuery {
	if query.Lang != "en" {
		query.Lang = "zh"
	} else {
		query.Lang = "en"
	}
	if query.Mode != "nsfw" {
		query.Mode = "sfw"
	} else {
		query.Mode = "nsfw"
	}
	if query.PageSize <= 0 {
		query.PageSize = 24
	}
	if query.PageSize > 60 {
		query.PageSize = 60
	}
	if query.Page <= 0 || query.Page > math.MaxInt64/query.PageSize {
		query.Page = 1
	}
	query.Q = strings.Clone(strings.TrimSpace(query.Q))
	if query.Phase == "" {
		query.Phase = "all"
	}
	if query.Sort == "" {
		query.Sort = "published_desc"
	}
	query.Phase = strings.Clone(query.Phase)
	query.Sort = strings.Clone(query.Sort)
	return query
}

func (s *CollectionService) metadata() v2models.CollectionMetadata {
	now := s.Now().UTC()
	return v2models.CollectionMetadata{SchemaVersion: 1, GeneratedAt: now, AsOfDate: now.Format(time.DateOnly)}
}

func collectionCacheKey(endpoint string, meta v2models.CollectionMetadata, query v2models.CollectionQuery) string {
	return "game:v2:collections:v3:" + endpoint + ":" + meta.AsOfDate + ":" + query.Lang + ":" + query.Mode
}

func (s *CollectionService) Home(ctx context.Context, query v2models.CollectionQuery) (v2models.CollectionHome, error) {
	query = NormalizeCollectionQuery(query)
	meta := s.metadata()
	key := collectionCacheKey("home", meta, query)
	return cachedCollection(ctx, s, key, func(value v2models.CollectionHome) bool {
		if !validCollectionMetadata(value.CollectionMetadata, meta) || value.Slots == nil {
			return false
		}
		// Retire heavy Home cache payloads in place; Index/Detail keep previews.
		for _, slot := range value.Slots {
			if slot.Collection.PreviewGames == nil || len(slot.Collection.PreviewGames) != 0 {
				return false
			}
		}
		return true
	}, func(ctx context.Context) (v2models.CollectionHome, error) {
		result := v2models.CollectionHome{CollectionMetadata: meta, Slots: []v2models.CollectionHomeSlot{}}
		records, err := s.reader.ListPublishedCollectionHomeSlots(ctx, query.Mode)
		if err != nil {
			return result, err
		}
		for _, record := range records {
			if record.VisibleGameCount == 0 {
				continue
			}
			result.Slots = append(result.Slots, v2models.CollectionHomeSlot{Slot: record.Slot, Collection: v2models.CollectionSummary{
				CollectionInfo: collectionInfo(record, query.Lang, int(record.VisibleGameCount)),
				PreviewGames:   []v2models.CollectionPreviewGame{},
			}})
		}
		return result, nil
	})
}

func (s *CollectionService) List(ctx context.Context, query v2models.CollectionQuery) (v2models.CollectionIndex, error) {
	query = NormalizeCollectionQuery(query)
	switch query.Phase {
	case "all", "released", "upcoming", "mixed":
	default:
		return v2models.CollectionIndex{}, ErrCollectionQuery
	}
	switch query.Sort {
	case "published_desc", "count_desc", "count_asc", "name_asc", "name_desc":
	default:
		return v2models.CollectionIndex{}, ErrCollectionQuery
	}
	meta := s.metadata()
	key := fmt.Sprintf("%s:%d:%d", collectionCacheKey("list", meta, query), query.Page, query.PageSize)
	build := func(ctx context.Context) (v2models.CollectionIndex, error) {
		result := v2models.CollectionIndex{CollectionMetadata: meta, Page: query.Page, PageSize: query.PageSize, Items: []v2models.CollectionSummary{}}
		total, err := s.reader.CountPublishedCollections(ctx, query)
		if err != nil {
			return result, err
		}
		result.Total = total
		result.HasMore = total > query.Page*query.PageSize
		records, err := s.reader.ListPublishedCollections(ctx, query)
		if err != nil {
			return result, err
		}
		timelines, err := s.timelines(ctx, records, query, meta.GeneratedAt)
		if err != nil {
			return result, err
		}
		for _, record := range records {
			result.Items = append(result.Items, collectionSummary(record, query.Lang, timelines[record.ID]))
		}
		return result, nil
	}
	if !query.IsDefaultBrowse() {
		// Arbitrary discovery criteria never enter the long-lived result cache.
		work, cancel := context.WithTimeout(ctx, 8*time.Second)
		defer cancel()
		return build(work)
	}
	return cachedCollection(ctx, s, key, func(value v2models.CollectionIndex) bool {
		return validCollectionMetadata(value.CollectionMetadata, meta) && value.Items != nil && value.Page == query.Page && value.PageSize == query.PageSize
	}, build)
}

func (s *CollectionService) Detail(ctx context.Context, code string, query v2models.CollectionQuery) (v2models.CollectionDetail, error) {
	if len(code) > 64 || !collectionCodePattern.MatchString(code) {
		return v2models.CollectionDetail{}, ErrCollectionNotFound
	}
	// A detached singleflight build must own Fiber's borrowed route string.
	code = strings.Clone(code)
	query = NormalizeCollectionQuery(query)
	meta := s.metadata()
	key := collectionCacheKey("detail", meta, query) + ":" + code
	return cachedCollection(ctx, s, key, func(value v2models.CollectionDetail) bool {
		return validCollectionMetadata(value.CollectionMetadata, meta) && value.Items != nil && value.Collection.Code == code
	}, func(ctx context.Context) (v2models.CollectionDetail, error) {
		result := v2models.CollectionDetail{CollectionMetadata: meta, Items: []v2models.CollectionTimelineItem{}}
		record, err := s.reader.GetPublishedCollection(ctx, code)
		if err != nil {
			return result, err
		}
		if record == nil {
			return result, ErrCollectionNotFound
		}
		timelines, err := s.timelines(ctx, []v2models.CollectionRecord{*record}, query, meta.GeneratedAt)
		if err != nil {
			return result, err
		}
		result.Items = timelines[record.ID]
		// Only visible members reach the decoration query. Preserve chronology order.
		if len(result.Items) > 0 {
			ids := make([]int64, len(result.Items))
			for i, item := range result.Items {
				ids[i], _ = strconv.ParseInt(item.GameID, 10, 64)
			}
			decorations, err := s.reader.LoadCollectionTimelineDecorations(ctx, ids, query.Lang)
			if err != nil {
				return result, err
			}
			for i, id := range ids {
				result.Items[i].CollectionTimelineDecoration = decorations[id]
			}
		}
		result.Collection = collectionInfo(*record, query.Lang, len(result.Items))
		return result, nil
	})
}

func (s *CollectionService) timelines(ctx context.Context, records []v2models.CollectionRecord, query v2models.CollectionQuery, asOfDate time.Time) (map[int64][]v2models.CollectionTimelineItem, error) {
	result := make(map[int64][]v2models.CollectionTimelineItem, len(records))
	ids := make([]int64, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
		result[record.ID] = []v2models.CollectionTimelineItem{}
	}
	if len(ids) == 0 {
		return result, nil
	}
	batch, err := s.reader.LoadCollectionProjectionGames(ctx, ids, query.Lang)
	if err != nil {
		return nil, err
	}
	games := make(map[int64]v2models.CollectionTimelineItem, len(batch.Games))
	for _, aggregate := range batch.Games {
		if query.Mode == "sfw" && aggregate.Adult {
			continue
		}
		// Share all existing V2 locale and media priorities, then expose only the
		// collection projection. No legacy maintenance source escapes this DTO.
		// The adapter reuses pure text fallback helpers, never the detail builder.
		text := v2models.GameV2Aggregate{Site: aggregate.Site, Details: aggregate.Details, Localized: aggregate.Localized}
		mediaLang := query.Lang
		if aggregate.Localized != nil && aggregate.Localized.Lang != "" {
			mediaLang = aggregate.Localized.Lang
		}
		header := canonicalGameHeader(aggregate.Site.Header, aggregate.Details, canonicalMediaHeader(aggregate.Media, aggregate.Assets, mediaLang))
		phase, chronology := resolveCollectionChronology(aggregate.FirstAvailable, aggregate.ReleaseState, asOfDate)
		games[aggregate.Site.ID] = v2models.CollectionTimelineItem{GameID: strconv.FormatInt(aggregate.Site.ID, 10), Name: localizedName(text, query.Lang), Summary: localizedSummary(text, query.Lang), HeaderURL: header, Phase: phase, Chronology: chronology}
	}
	for _, member := range batch.Memberships {
		if game, ok := games[member.GameID]; ok {
			result[member.CollectionID] = append(result[member.CollectionID], game)
		}
	}
	for _, items := range result {
		sortCollectionTimeline(items)
	}
	return result, nil
}

func collectionInfo(record v2models.CollectionRecord, lang string, count int) v2models.CollectionInfo {
	name, info := record.Name, record.Info
	fallbackName, fallbackInfo := record.NameEn, record.InfoEn
	if lang == "en" {
		name, info, fallbackName, fallbackInfo = fallbackName, fallbackInfo, name, info
	}
	if strings.TrimSpace(name) == "" {
		name = fallbackName
	}
	if strings.TrimSpace(info) == "" {
		info = fallbackInfo
	}
	return v2models.CollectionInfo{Code: record.Code, Name: name, Info: info, VisibleGameCount: count, PublishedAt: record.PublishedAt.UTC()}
}

func collectionSummary(record v2models.CollectionRecord, lang string, items []v2models.CollectionTimelineItem) v2models.CollectionSummary {
	return v2models.CollectionSummary{CollectionInfo: collectionInfo(record, lang, len(items)), PreviewGames: collectionPreview(items)}
}

func validCollectionMetadata(value, expected v2models.CollectionMetadata) bool {
	return value.SchemaVersion == 1 && value.AsOfDate == expected.AsOfDate && !value.GeneratedAt.IsZero()
}

func cachedCollection[T any](ctx context.Context, s *CollectionService, key string, valid func(T) bool, build func(context.Context) (T, error)) (T, error) {
	var zero T
	load := func(ctx context.Context) (T, bool) {
		var value T
		if s.cache == nil {
			return value, false
		}
		bounded, cancel := context.WithTimeout(ctx, collectionRedisTimeout)
		defer cancel()
		data, err := s.cache.Get(bounded, key).Bytes()
		return valueFromCollectionCache(data, err, valid)
	}
	if value, ok := load(ctx); ok {
		return value, nil
	}
	// A canceled waiter does not poison the other same-key readers. Work has its
	// own finite budget; each waiter still observes its own cancellation.
	flight := s.flights.DoChan(key, func() (any, error) {
		work, cancel := context.WithTimeout(context.WithoutCancel(ctx), 8*time.Second)
		defer cancel()
		if value, ok := load(work); ok {
			return json.Marshal(value)
		}
		value, err := build(work)
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		if s.cache != nil {
			bounded, stop := context.WithTimeout(work, collectionRedisTimeout)
			_ = s.cache.Set(bounded, key, data, CollectionCacheTTL).Err()
			stop()
		}
		return data, nil
	})
	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	case result := <-flight:
		if result.Err != nil {
			return zero, result.Err
		}
		// Decode per caller: a shared flight must not share mutable DTO slices.
		var value T
		err := json.Unmarshal(result.Val.([]byte), &value)
		return value, err
	}
}

func valueFromCollectionCache[T any](data []byte, err error, valid func(T) bool) (T, bool) {
	var value T
	if err != nil || json.Unmarshal(data, &value) != nil {
		return value, false
	}
	return value, valid(value)
}
