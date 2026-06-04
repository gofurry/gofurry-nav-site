package service

import (
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	navdao "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/dao"
	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/siteindex/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type siteIndexStore interface {
	GetSiteIndexList() ([]navmodels.GfnSiteIndex, common.GFError)
}

type siteIndexService struct {
	store siteIndexStore
	now   func() time.Time
}

var (
	siteIndexSingleton = &siteIndexService{}
	siteIndexMu        sync.Mutex
)

func GetSiteIndexService() *siteIndexService {
	siteIndexMu.Lock()
	defer siteIndexMu.Unlock()
	if siteIndexSingleton.store == nil {
		siteIndexSingleton.store = navdao.GetNavPageDao()
	}
	if siteIndexSingleton.now == nil {
		siteIndexSingleton.now = time.Now
	}
	return siteIndexSingleton
}

func newSiteIndexService(store siteIndexStore, now func() time.Time) *siteIndexService {
	return &siteIndexService{store: store, now: now}
}

func (svc *siteIndexService) GetSiteIndex() models.SiteIndexResponse {
	response := models.SiteIndexResponse{
		SchemaVersion: models.SiteIndexSchemaVersion,
		GeneratedAt:   svc.clock()(),
		State:         models.SiteIndexStateEmpty,
		Items:         []models.SiteIndexItem{},
	}

	records, err := svc.source().GetSiteIndexList()
	if err != nil {
		response.State = models.SiteIndexStateError
		response.ReasonMessages = []string{err.GetMsg()}
		return response
	}
	if len(records) == 0 {
		return response
	}

	response.State = models.SiteIndexStateReady
	response.Items = make([]models.SiteIndexItem, 0, len(records))
	for _, record := range records {
		response.Items = append(response.Items, models.SiteIndexItem{
			ID:        record.ID,
			Domains:   parseDomains(record.Domain),
			UpdatedAt: record.UpdateTime,
		})
	}
	return response
}

func (svc *siteIndexService) source() siteIndexStore {
	if svc != nil && svc.store != nil {
		return svc.store
	}
	return navdao.GetNavPageDao()
}

func (svc *siteIndexService) clock() func() time.Time {
	if svc != nil && svc.now != nil {
		return svc.now
	}
	return time.Now
}

func parseDomains(raw string) []string {
	type domainEnvelope struct {
		Domain []string `json:"domain"`
	}
	var envelope domainEnvelope
	if err := sonic.Unmarshal([]byte(raw), &envelope); err == nil && len(envelope.Domain) > 0 {
		return sanitizeDomains(envelope.Domain)
	}

	var list []string
	if err := sonic.Unmarshal([]byte(raw), &list); err == nil && len(list) > 0 {
		return sanitizeDomains(list)
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	return []string{raw}
}

func sanitizeDomains(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}
