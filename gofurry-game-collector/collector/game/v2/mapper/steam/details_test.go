package steam

import (
	"testing"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/steam-go/web/storefront"
)

func TestDetailsMapperMapsHighValueFields(t *testing.T) {
	t.Parallel()

	data := storefront.AppDetailsData{
		Type:               "game",
		Name:               "Fixture Game",
		IsFree:             false,
		Website:            "https://example.test",
		HeaderImage:        "header.jpg",
		CapsuleImage:       "capsule.jpg",
		SupportedLanguages: "English",
		Developers:         []string{"Dev"},
		Publishers:         []string{"Pub"},
		ReleaseDate:        &storefront.StoreReleaseDate{ComingSoon: false, Date: "1 Jan, 2026"},
		Platforms:          storefront.StorePlatforms{Windows: true},
		SupportInfo:        &storefront.StoreSupportInfo{URL: "https://support.test"},
		ContentDescriptors: &storefront.StoreContentDescriptors{IDs: []int{2}, Notes: "notes"},
		Ratings:            []byte(`{"steam_germany":{"required_age":"18"},"usk":{"rating":"16"}}`),
		PriceOverview:      &storefront.StorePrice{Currency: "USD", Initial: 999, Final: 499, DiscountPercent: 50, FinalFormatted: "$4.99"},
		Screenshots:        []storefront.StoreScreenshot{{ID: 1, PathFull: "full.jpg"}},
		Movies:             []storefront.StoreMovie{{ID: 2, Name: "Trailer", DASHH264: "movie.mpd"}},
		PCRequirements:     &storefront.StoreRequirements{Minimum: "min"},
	}

	mapper := NewDetailsMapper()
	now := time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC)
	details, err := mapper.ToDetails(10, 550, data, now)
	if err != nil {
		t.Fatalf("ToDetails returned error: %v", err)
	}
	if details.Name != "Fixture Game" || details.Ratings[0].Board == "" {
		t.Fatalf("unexpected details: %#v", details)
	}
	price := mapper.ToPrice(10, 550, domain.RegionUS, data, now)
	if price.Final != 499 || price.Currency != "USD" {
		t.Fatalf("unexpected price: %#v", price)
	}
	media := mapper.ToMedia(10, 550, data, now)
	if len(media.Screenshots) != 1 || len(media.Movies) != 1 || media.Movies[0].DASHH264URL != "movie.mpd" {
		t.Fatalf("unexpected media: %#v", media)
	}
	requirements := mapper.ToRequirements(10, 550, data, now)
	if requirements.PC.Minimum != "min" {
		t.Fatalf("unexpected requirements: %#v", requirements)
	}
}
