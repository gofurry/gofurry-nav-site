package steam

import (
	"testing"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/steam-go/addons/assets"
)

func TestDetailsMapperKeepsOfficialStoreBrowseAssetURLs(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	items := []assets.URLItem{
		{
			AppID:    550,
			Kind:     assets.KindHeader,
			URL:      "https://shared.steamstatic.com/store_item_assets/steam/apps/550/header.jpg?t=1",
			Filename: "header.jpg",
			Source:   assets.SourceStoreBrowse,
		},
		{
			AppID:    550,
			Kind:     assets.KindLibraryCapsule2x,
			URL:      "https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/550/0508d712ca859e3ef0921a32c49088ccfb05b0a0/library_capsule_2x.jpg?t=2",
			Digest:   "0508d712ca859e3ef0921a32c49088ccfb05b0a0",
			Filename: "library_capsule_2x.jpg",
			Source:   assets.SourceStoreBrowse,
		},
	}

	got := NewDetailsMapper().ToStoreBrowseAssets(10, 550, domain.LanguageEN, items, now)
	if len(got) != 2 {
		t.Fatalf("asset count = %d, want 2: %#v", len(got), got)
	}
	if got[0].URL != "https://shared.steamstatic.com/store_item_assets/steam/apps/550/header.jpg" {
		t.Fatalf("legacy official URL was not preserved and cleaned: %s", got[0].URL)
	}
	if got[1].URL != "https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/550/0508d712ca859e3ef0921a32c49088ccfb05b0a0/library_capsule_2x.jpg" {
		t.Fatalf("hashed official URL was not preserved and cleaned: %s", got[1].URL)
	}
	if got[0].Source != assets.SourceStoreBrowse || got[1].Source != assets.SourceStoreBrowse {
		t.Fatalf("unexpected source: %#v", got)
	}
}
