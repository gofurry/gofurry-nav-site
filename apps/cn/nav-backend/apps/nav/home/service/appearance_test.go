package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/home/models"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5"
)

type fakeAppearance struct {
	// Unused catalog methods remain available to small pool-only fixtures.
	appearanceReader
	desktop, mobile bool
	failed          string
}

func (f fakeAppearance) RandomHeroAsset(_ context.Context, variant string) (navsqlc.RandomHeroAssetRow, error) {
	if f.failed == variant {
		return navsqlc.RandomHeroAssetRow{}, errors.New("unavailable")
	}
	if (variant == "desktop" && !f.desktop) || (variant == "mobile" && !f.mobile) {
		return navsqlc.RandomHeroAssetRow{}, pgx.ErrNoRows
	}
	return navsqlc.RandomHeroAssetRow{ID: 9007199254740993, ObjectKey: "nav/hero/" + variant + "/" + strings.Repeat("a", 32) + ".avif"}, nil
}
func (fakeAppearance) PublicBackgroundPatterns(context.Context) ([]navsqlc.PublicBackgroundPatternsRow, error) {
	return []navsqlc.PublicBackgroundPatternsRow{{ID: 9007199254740993, Name: "图案", NameEn: "Pattern", ObjectKey: "nav/patterns/" + strings.Repeat("b", 32) + ".svg", DefaultSizePx: 96, LightOpacity: 0.1}}, nil
}
func TestIndependentHeroPools(t *testing.T) {
	for _, tc := range []fakeAppearance{{desktop: true, mobile: true}, {desktop: true}, {mobile: true}, {}, {desktop: true, mobile: true, failed: "desktop"}} {
		svc := newHomeService(nil, time.Now).WithAppearance(tc)
		result := svc.GetHomeHero(context.Background(), models.HeroSelection{})
		if (result.Hero.Desktop != nil) != (tc.desktop && tc.failed != "desktop") || (result.Hero.Mobile != nil) != (tc.mobile && tc.failed != "mobile") {
			t.Fatalf("pool crossed or partial success lost: %+v", result)
		}
		data, _ := json.Marshal(result)
		if strings.Contains(string(data), "https://") || strings.Contains(string(data), "backgrounds") {
			t.Fatal("legacy asset representation leaked")
		}
		if result.Hero.Desktop != nil && !strings.Contains(string(data), `"id":"9007199254740993"`) {
			t.Fatal("unsafe numeric ID")
		}
	}
}

func (f fakeAppearance) PublicHeroAsset(_ context.Context, query navsqlc.PublicHeroAssetParams) (navsqlc.PublicHeroAssetRow, error) {
	if f.failed == query.Variant {
		return navsqlc.PublicHeroAssetRow{}, errors.New("unavailable")
	}
	if (query.ID == 10 && query.Variant == "desktop") || (query.ID == 20 && query.Variant == "mobile") {
		return navsqlc.PublicHeroAssetRow{ID: query.ID, ObjectKey: "nav/hero/" + query.Variant + "/" + strings.Repeat("b", 32) + ".avif"}, nil
	}
	return navsqlc.PublicHeroAssetRow{}, pgx.ErrNoRows
}

func TestFixedHeroSelection(t *testing.T) {
	for _, tc := range []struct {
		selection       models.HeroSelection
		desktop, mobile int64
	}{
		{models.HeroSelection{DesktopID: 10, MobileID: 20}, 10, 20},
		{models.HeroSelection{DesktopID: 999, MobileID: 20}, 9007199254740993, 20},
		{models.HeroSelection{DesktopID: 20, MobileID: 10}, 9007199254740993, 9007199254740993},
		{models.HeroSelection{DesktopID: 10}, 10, 9007199254740993},
	} {
		hero := newHomeService(nil, time.Now).WithAppearance(fakeAppearance{desktop: true, mobile: true}).GetHomeHero(context.Background(), tc.selection).Hero
		if hero.Desktop.ID != tc.desktop || hero.Mobile.ID != tc.mobile {
			t.Fatalf("independent fixed selection failed: %+v", hero)
		}
	}
	hero := newHomeService(nil, time.Now).WithAppearance(fakeAppearance{desktop: true, mobile: true, failed: "desktop"}).GetHomeHero(context.Background(), models.HeroSelection{DesktopID: 10, MobileID: 20}).Hero
	if hero.Desktop != nil || hero.Mobile == nil || hero.Mobile.ID != 20 {
		t.Fatal("fixed read failure lost other viewport")
	}
	// Nil reader deliberately proves Local never performs any cloud query.
	local := newHomeService(nil, time.Now).GetHomeHero(context.Background(), models.HeroSelection{Local: true})
	if local.Hero.Desktop != nil || local.Hero.Mobile != nil || local.State != models.HomeStateReady {
		t.Fatal("local selected cloud assets")
	}
}
func TestPatternPublicContract(t *testing.T) {
	svc := newHomeService(nil, time.Now).WithAppearance(fakeAppearance{})
	result, err := svc.GetPatterns(context.Background())
	if err != nil || result.SchemaVersion != 1 || len(result.Patterns) != 1 {
		t.Fatal("catalog missing")
	}
	data, _ := json.Marshal(result)
	if !strings.Contains(string(data), `"id":"9007199254740993"`) || strings.Contains(string(data), "https://") {
		t.Fatal("public asset contract violated")
	}
	if models.HomeSchemaVersion != 4 {
		t.Fatal("home schema was not bumped")
	}
}
