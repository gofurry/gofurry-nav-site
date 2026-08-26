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

func TestDetailsMapperMapsCanonicalReleaseAndLanguages(t *testing.T) {
	t.Parallel()

	mapper := NewDetailsMapper()
	observedAt := time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)
	release := mapper.ToCanonicalRelease(10, &storefront.StoreReleaseDate{ComingSoon: true, Date: "Q3 2026"}, observedAt)
	if release.Availability != domain.ReleaseUpcoming || release.Precision != domain.ReleasePrecisionQuarter || release.Year == nil || *release.Year != 2026 || release.Quarter == nil || *release.Quarter != 3 {
		t.Fatalf("unexpected canonical release: %#v", release)
	}
	if release.WindowStart == nil || release.WindowStart.Format(time.DateOnly) != "2026-07-01" || release.WindowEnd == nil || release.WindowEnd.Format(time.DateOnly) != "2026-09-30" {
		t.Fatalf("unexpected canonical release window: %#v", release)
	}

	parsed := storefront.ParseSupportedLanguages(`English<strong>*</strong>, Klingon<br><strong>*</strong>languages with full audio support`)
	languages := mapper.ToCanonicalLanguages(parsed, observedAt)
	if len(languages.Items) != 2 || languages.Items[0].Code == nil || *languages.Items[0].Code != "en" || languages.Items[0].FullAudioSupported == nil || !*languages.Items[0].FullAudioSupported {
		t.Fatalf("unexpected known language: %#v", languages.Items)
	}
	if languages.Items[1].Code != nil || languages.Items[1].SteamName != "Klingon" {
		t.Fatalf("unknown language was not preserved: %#v", languages.Items[1])
	}
}

func TestDetailsMapperCanonicalReleasePrecisions(t *testing.T) {
	t.Parallel()
	mapper := NewDetailsMapper()
	for name, test := range map[string]struct {
		value     storefront.StoreReleaseDate
		precision domain.ReleasePrecision
		raw       string
	}{
		"day":     {storefront.StoreReleaseDate{Date: "24 Aug, 2026"}, domain.ReleasePrecisionDay, "24 Aug, 2026"},
		"month":   {storefront.StoreReleaseDate{ComingSoon: true, Date: "August 2026"}, domain.ReleasePrecisionMonth, "August 2026"},
		"quarter": {storefront.StoreReleaseDate{ComingSoon: true, Date: "Q3 2026"}, domain.ReleasePrecisionQuarter, "Q3 2026"},
		"year":    {storefront.StoreReleaseDate{ComingSoon: true, Date: "2026"}, domain.ReleasePrecisionYear, "2026"},
		"tba":     {storefront.StoreReleaseDate{ComingSoon: true, Date: "Coming Soon"}, domain.ReleasePrecisionTBA, "Coming Soon"},
		"none":    {storefront.StoreReleaseDate{}, domain.ReleasePrecisionNone, ""},
		"unknown": {storefront.StoreReleaseDate{ComingSoon: true, Date: "Fall 2026"}, domain.ReleasePrecisionUnknown, "Fall 2026"},
	} {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := mapper.ToCanonicalRelease(1, &test.value, time.Now())
			if got.Precision != test.precision || got.RawText != test.raw || got.SourceRegion != domain.RegionUS || got.SourceLocale != domain.StoreLocaleEN || got.Normalizer != "steam-go/v1.3.9" {
				t.Fatalf("unexpected canonical release: %#v", got)
			}
			if test.precision == domain.ReleasePrecisionUnknown && (got.WindowStart != nil || got.Year != nil) {
				t.Fatalf("unknown release carried fabricated calendar fields: %#v", got)
			}
		})
	}
}
