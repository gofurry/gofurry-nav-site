package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/home/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5"
)

type appearanceReader interface {
	RandomHeroAsset(context.Context, string) (navsqlc.RandomHeroAssetRow, error)
	PublicBackgroundPatterns(context.Context) ([]navsqlc.PublicBackgroundPatternsRow, error)
}

func (svc *homeService) WithAppearance(reader appearanceReader) *homeService {
	svc.appearance = reader
	return svc
}

// Hero selection bypasses the derived home cache. Disabled/deleted assets stop
// appearing immediately, and each viewport pool is sampled independently.
func (svc *homeService) GetHomeHero(ctx context.Context) models.HomeHeroResponse {
	response := models.HomeHeroResponse{SchemaVersion: models.HomeSchemaVersion, GeneratedAt: svc.clock()(), State: models.HomeStateReady}
	if svc.appearance == nil {
		response.State = models.HomeStateMissing
		return response
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var desktop, mobile *models.HeroAsset
	var desktopErr, mobileErr error
	read := func(variant string) (*models.HeroAsset, error) {
		row, err := svc.appearance.RandomHeroAsset(ctx, variant)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		return &models.HeroAsset{ID: row.ID, ObjectKey: row.ObjectKey}, nil
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); desktop, desktopErr = read("desktop") }()
	go func() { defer wg.Done(); mobile, mobileErr = read("mobile") }()
	wg.Wait()
	response.Hero = models.HomeHero{Desktop: desktop, Mobile: mobile}
	if desktopErr != nil {
		response.ReasonMessages = append(response.ReasonMessages, "Desktop hero pool unavailable")
	}
	if mobileErr != nil {
		response.ReasonMessages = append(response.ReasonMessages, "Mobile hero pool unavailable")
	}
	if desktop == nil && mobile == nil {
		response.State = models.HomeStateMissing
	}
	return response
}

func (svc *homeService) GetPatterns(ctx context.Context) (models.PatternCatalog, common.GFError) {
	response := models.PatternCatalog{SchemaVersion: 1, Patterns: []models.BackgroundPattern{}}
	if svc.appearance == nil {
		return response, common.NewServiceError("Pattern catalog unavailable")
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	rows, err := svc.appearance.PublicBackgroundPatterns(ctx)
	if err != nil {
		return response, common.NewServiceError("Pattern catalog unavailable")
	}
	for _, row := range rows {
		response.Patterns = append(response.Patterns, models.BackgroundPattern{ID: row.ID, Name: row.Name, NameEn: row.NameEn, ObjectKey: row.ObjectKey, LightColor: row.LightColor, DarkColor: row.DarkColor, LightOpacity: row.LightOpacity, DarkOpacity: row.DarkOpacity, DefaultSizePx: row.DefaultSizePx})
	}
	return response, nil
}
