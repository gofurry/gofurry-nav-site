package service

import (
	"strings"
	"sync"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/updates/models"
	database "github.com/gofurry/gofurry-nav-backend/roof/db"
	"gorm.io/gorm"
)

const defaultUpdatesLimit = 100

type updateNoticeStore interface {
	ListUpdateNotices(limit int) ([]models.UpdateNotice, error)
}

type gormUpdateNoticeStore struct {
	db *gorm.DB
}

func newGormUpdateNoticeStore() *gormUpdateNoticeStore {
	return &gormUpdateNoticeStore{db: database.Orm.DB()}
}

func (store *gormUpdateNoticeStore) ListUpdateNotices(limit int) ([]models.UpdateNotice, error) {
	if limit <= 0 {
		limit = defaultUpdatesLimit
	}
	var notices []models.UpdateNotice
	err := store.db.Model(&models.UpdateNotice{}).
		Where("deleted IS NOT TRUE").
		Order("published_at DESC, id DESC").
		Limit(limit).
		Find(&notices).Error
	return notices, err
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
	if updatesSingleton.store == nil {
		updatesSingleton.store = newGormUpdateNoticeStore()
	}
	if updatesSingleton.now == nil {
		updatesSingleton.now = time.Now
	}
	return updatesSingleton
}

func newUpdatesService(store updateNoticeStore, now func() time.Time) *updatesService {
	return &updatesService{store: store, now: now}
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
	if svc != nil && svc.store != nil {
		return svc.store
	}
	return newGormUpdateNoticeStore()
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
