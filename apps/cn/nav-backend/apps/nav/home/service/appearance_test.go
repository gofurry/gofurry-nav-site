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
		result := svc.GetHomeHero(context.Background())
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
