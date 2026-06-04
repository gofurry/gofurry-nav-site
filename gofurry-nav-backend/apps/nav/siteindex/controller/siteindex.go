package controller

import (
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/siteindex/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/siteindex/service"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type siteIndexReader interface {
	GetSiteIndex() models.SiteIndexResponse
}

type siteIndexApi struct{}

var SiteIndexApi *siteIndexApi

var (
	siteIndexReaderMu      sync.RWMutex
	siteIndexReaderForTest siteIndexReader
)

func init() {
	SiteIndexApi = &siteIndexApi{}
}

func (api siteIndexApi) GetSiteIndex(c fiber.Ctx) error {
	return common.NewResponse(c).SuccessWithData(currentSiteIndexReader().GetSiteIndex())
}

func currentSiteIndexReader() siteIndexReader {
	siteIndexReaderMu.RLock()
	reader := siteIndexReaderForTest
	siteIndexReaderMu.RUnlock()
	if reader != nil {
		return reader
	}
	return service.GetSiteIndexService()
}

func setSiteIndexReaderForTest(reader siteIndexReader) func() {
	siteIndexReaderMu.Lock()
	previous := siteIndexReaderForTest
	siteIndexReaderForTest = reader
	siteIndexReaderMu.Unlock()
	return func() {
		siteIndexReaderMu.Lock()
		siteIndexReaderForTest = previous
		siteIndexReaderMu.Unlock()
	}
}
