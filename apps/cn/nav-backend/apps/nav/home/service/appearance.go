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
	PublicHeroAsset(context.Context, navsqlc.PublicHeroAssetParams) (navsqlc.PublicHeroAssetRow, error)
	CountPublicHeroAssets(context.Context, string) (int64, error)
	PublicHeroAssets(context.Context, navsqlc.PublicHeroAssetsParams) ([]navsqlc.PublicHeroAssetsRow, error)
	PublicBackgroundPatterns(context.Context) ([]navsqlc.PublicBackgroundPatternsRow, error)
}

func (svc *homeService) WithAppearance(reader appearanceReader) *homeService {
	svc.appearance = reader
	return svc
}

// Hero selection bypasses the derived home cache. Disabled/deleted assets stop
// appearing immediately, and each viewport pool is sampled independently.
func (svc *homeService) GetHomeHero(ctx context.Context, selection models.HeroSelection) models.HomeHeroResponse {
	response := models.HomeHeroResponse{SchemaVersion: models.HomeSchemaVersion, GeneratedAt: svc.clock()(), State: models.HomeStateReady}
	if selection.Local {
		return response
	}
	if svc.appearance == nil {
		response.State = models.HomeStateMissing
		return response
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var desktop, mobile *models.HeroAsset
	var desktopErr, mobileErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); desktop, desktopErr = svc.resolveHero(ctx, "desktop", selection.DesktopID) }()
	go func() { defer wg.Done(); mobile, mobileErr = svc.resolveHero(ctx, "mobile", selection.MobileID) }()
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

func (svc *homeService) resolveHero(ctx context.Context, variant string, id int64) (*models.HeroAsset, error) {
	if id > 0 {
		row, err := svc.appearance.PublicHeroAsset(ctx, navsqlc.PublicHeroAssetParams{ID: id, Variant: variant})
		if err == nil {
			return &models.HeroAsset{ID: row.ID, ObjectKey: row.ObjectKey}, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}
	row, err := svc.appearance.RandomHeroAsset(ctx, variant)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &models.HeroAsset{ID: row.ID, ObjectKey: row.ObjectKey}, nil
}

func (svc *homeService) GetHeroes(ctx context.Context, query models.HeroCatalogQuery) (models.HeroCatalog, common.GFError) {
	response := models.HeroCatalog{SchemaVersion: 1, Variant: query.Variant, PageNum: query.PageNum, PageSize: query.PageSize, Items: []models.HeroCatalogItem{}}
	unavailable := common.NewServiceError("Hero catalog unavailable")
	if svc.appearance == nil {
		return response, unavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	total, err := svc.appearance.CountPublicHeroAssets(ctx, query.Variant)
	if err != nil {
		return response, unavailable
	}
	response.Total = total
	rows, err := svc.appearance.PublicHeroAssets(ctx, navsqlc.PublicHeroAssetsParams{Variant: query.Variant, PageSize: query.PageSize, PageOffset: (query.PageNum - 1) * query.PageSize})
	if err != nil {
		return response, unavailable
	}
	for _, row := range rows {
		response.Items = append(response.Items, models.HeroCatalogItem{ID: row.ID, Name: row.Name, ObjectKey: row.ObjectKey})
	}
	if query.SelectedID > 0 {
		row, err := svc.appearance.PublicHeroAsset(ctx, navsqlc.PublicHeroAssetParams{ID: query.SelectedID, Variant: query.Variant})
		if err == nil {
			response.Selected = &models.HeroCatalogItem{ID: row.ID, Name: row.Name, ObjectKey: row.ObjectKey}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return response, unavailable
		}
	}
	return response, nil
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
