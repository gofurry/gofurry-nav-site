package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/updates/models"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
)

const defaultUpdatesLimit = 100

type updateNoticeStore interface {
	ListUpdateNotices(limit int) ([]models.UpdateNotice, error)
}

type sqlcUpdateNoticeStore struct {
	queries *navsqlc.Queries
}

func newSQLCUpdateNoticeStore(queries *navsqlc.Queries) *sqlcUpdateNoticeStore {
	return &sqlcUpdateNoticeStore{queries: queries}
}

func (store *sqlcUpdateNoticeStore) ListUpdateNotices(limit int) ([]models.UpdateNotice, error) {
	if limit <= 0 {
		limit = defaultUpdatesLimit
	}
	rows, err := store.queries.ListPublicUpdateNotices(context.Background(), int32(limit))
	if err != nil {
		return nil, err
	}
	notices := make([]models.UpdateNotice, 0, len(rows))
	for _, row := range rows {
		notices = append(notices, models.UpdateNotice{
			ID: row.ID, Title: row.Title, TitleEn: row.TitleEn, Body: row.Body, BodyEn: row.BodyEn,
			PublishedAt: row.PublishedAt.Time, CreateTime: row.CreateTime.Time,
			UpdateTime: row.UpdateTime.Time, Deleted: row.Deleted,
		})
	}
	return notices, nil
}

type updatesService struct {
	store updateNoticeStore
	now   func() time.Time
}

var (
	updatesSingleton = &updatesService{}
	updatesMu        sync.Mutex
)

func GetUpdatesService() *updatesService {
	updatesMu.Lock()
	defer updatesMu.Unlock()
	if updatesSingleton.now == nil {
		updatesSingleton.now = time.Now
	}
	return updatesSingleton
}

func newUpdatesService(store updateNoticeStore, now func() time.Time) *updatesService {
	return &updatesService{store: store, now: now}
}

func New(queries *navsqlc.Queries) *updatesService {
	return newUpdatesService(newSQLCUpdateNoticeStore(queries), time.Now)
}

func (svc *updatesService) GetUpdates(lang string) models.UpdatesResponse {
	lang = normalizeLang(lang)
	response := models.UpdatesResponse{
		SchemaVersion: models.UpdatesSchemaVersion,
		GeneratedAt:   svc.clock()(),
		State:         models.UpdatesStateEmpty,
		Items:         []models.UpdateNoticeItem{},
	}

	notices, err := svc.source().ListUpdateNotices(defaultUpdatesLimit)
	if err != nil {
		response.State = models.UpdatesStateError
		response.ReasonMessages = []string{err.Error()}
		return response
	}

	if len(notices) == 0 {
		return response
	}

	response.State = models.UpdatesStateReady
	response.Items = make([]models.UpdateNoticeItem, 0, len(notices))
	for _, notice := range notices {
		title, body := localizeNotice(notice, lang)
		response.Items = append(response.Items, models.UpdateNoticeItem{
			ID:          notice.ID,
			Title:       title,
			Body:        body,
			PublishedAt: notice.PublishedAt,
			CreateTime:  notice.CreateTime,
			UpdateTime:  notice.UpdateTime,
		})
	}
	return response
}

func (svc *updatesService) source() updateNoticeStore {
	return svc.store
}

func (svc *updatesService) clock() func() time.Time {
	if svc != nil && svc.now != nil {
		return svc.now
	}
	return time.Now
}

func normalizeLang(lang string) string {
	if strings.EqualFold(strings.TrimSpace(lang), "en") {
		return "en"
	}
	return "zh"
}

func localizeNotice(notice models.UpdateNotice, lang string) (string, string) {
	if lang == "en" {
		return firstNonEmpty(notice.TitleEn, notice.Title), firstNonEmpty(notice.BodyEn, notice.Body)
	}
	return firstNonEmpty(notice.Title, notice.TitleEn), firstNonEmpty(notice.Body, notice.BodyEn)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
