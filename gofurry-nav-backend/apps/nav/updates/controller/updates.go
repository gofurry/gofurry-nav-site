package controller

import (
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/updates/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/updates/service"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type updatesApi struct{}

var UpdatesApi *updatesApi

func init() {
	UpdatesApi = &updatesApi{}
}

type updatesReader interface {
	GetUpdates(lang string) models.UpdatesResponse
}

var (
	updatesReaderMu      sync.RWMutex
	updatesReaderForTest updatesReader
)

func (api updatesApi) GetUpdates(c fiber.Ctx) error {
	data := currentUpdatesReader().GetUpdates(c.Query("lang", "zh"))
	return common.NewResponse(c).SuccessWithData(data)
}

func currentUpdatesReader() updatesReader {
	updatesReaderMu.RLock()
	reader := updatesReaderForTest
	updatesReaderMu.RUnlock()
	if reader != nil {
		return reader
	}
	return service.GetUpdatesService()
}

func setUpdatesReaderForTest(reader updatesReader) func() {
	updatesReaderMu.Lock()
	previous := updatesReaderForTest
	updatesReaderForTest = reader
	updatesReaderMu.Unlock()
	return func() {
		updatesReaderMu.Lock()
		updatesReaderForTest = previous
		updatesReaderMu.Unlock()
	}
}
