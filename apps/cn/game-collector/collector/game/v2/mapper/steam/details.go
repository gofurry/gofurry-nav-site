package steam

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/steam-go/web/storefront"
)

const steamNormalizerVersion = "steam-go/v1.3.9"

// DetailsMapper converts steam-go Store appdetails into collector v2 domain models.
type DetailsMapper struct{}

// NewDetailsMapper returns a details mapper.
func NewDetailsMapper() DetailsMapper {
	return DetailsMapper{}
}

// ToDetails maps the cross-language details subset.
func (m DetailsMapper) ToDetails(gameID int64, appID uint32, data storefront.AppDetailsData, collectedAt time.Time) (domain.GameDetails, error) {
	ratings, err := mapRatings(data)
	if err != nil {
		return domain.GameDetails{}, err
	}
	return domain.GameDetails{
		GameID:                gameID,
		AppID:                 appID,
		Type:                  data.Type,
		Name:                  data.Name,
		IsFree:                data.IsFree,
		Website:               data.Website,
		HeaderURL:             data.HeaderImage,
		Developers:            append([]string(nil), data.Developers...),
		Publishers:            append([]string(nil), data.Publishers...),
		ReleaseRaw:            mapReleaseDate(data.ReleaseDate),
		Platforms:             mapPlatforms(data.Platforms),
		SupportedLanguagesRaw: data.SupportedLanguages,
		SupportInfo:           mapSupportInfo(data.SupportInfo),
		ContentDescriptors:    mapContentDescriptors(data.ContentDescriptors),
		Ratings:               ratings,
		CollectedAt:           collectedAt,
	}, nil
}

// ToLocalized maps language-specific copy.
func (m DetailsMapper) ToLocalized(gameID int64, appID uint32, lang domain.StoreLocale, data storefront.AppDetailsData, collectedAt time.Time) domain.GameLocalizedDetails {
	return domain.GameLocalizedDetails{
		GameID:              gameID,
		AppID:               appID,
		Language:            lang,
		Name:                data.Name,
		ShortDescription:    data.ShortDescription,
		DetailedDescription: data.DetailedDescription,
		AboutTheGame:        data.AboutTheGame,
		CollectedAt:         collectedAt,
	}
}

// ToCanonicalRelease maps one authoritative US/English release observation.
// Callers must only invoke it for a non-nil Storefront release_date field.
func (m DetailsMapper) ToCanonicalRelease(gameID int64, value *storefront.StoreReleaseDate, observedAt time.Time) domain.GameReleaseState {
	normalized := storefront.NormalizeReleaseDate(value)
	availability := domain.ReleaseAvailable
	if normalized.ComingSoon {
		availability = domain.ReleaseUpcoming
	}
	return domain.GameReleaseState{
		GameID:       gameID,
		Availability: availability,
		Precision:    domain.ReleasePrecision(normalized.Precision),
		ExactDate:    cloneDate(normalized.ExactDate),
		Year:         positiveInt(normalized.Year),
		Month:        positiveInt(normalized.Month),
		Quarter:      positiveInt(normalized.Quarter),
		WindowStart:  cloneDate(normalized.RangeStart),
		WindowEnd:    cloneDate(normalized.RangeEnd),
		RawText:      normalized.RawText,
		Source:       domain.SourceSteam,
		SourceRegion: domain.RegionUS,
		SourceLocale: domain.StoreLocaleEN,
		Normalizer:   steamNormalizerVersion,
		ObservedAt:   observedAt,
	}
}

// ToCanonicalLanguages maps the ordered language set parsed by steam-go.
func (m DetailsMapper) ToCanonicalLanguages(value []storefront.LanguageSupport, observedAt time.Time) domain.GameLanguages {
	items := make([]domain.GameLanguage, 0, len(value))
	for index, support := range value {
		var code *string
		if support.Known {
			code = nonEmptyString(support.Code)
		}
		items = append(items, domain.GameLanguage{
			Code:               code,
			SteamName:          support.SteamName,
			SteamAPICode:       nonEmptyString(support.SteamAPICode),
			SteamWebCode:       nonEmptyString(support.SteamWebCode),
			Tier:               string(support.Tier),
			InterfaceSupported: support.Interface,
			SubtitlesSupported: support.Subtitles,
			FullAudioSupported: support.FullAudio,
			SortOrder:          index,
			Source:             domain.SourceSteam,
			SourceRegion:       domain.RegionUS,
			SourceLocale:       domain.StoreLocaleEN,
			Normalizer:         steamNormalizerVersion,
			ObservedAt:         observedAt,
		})
	}
	return domain.GameLanguages{Items: items}
}

// ToPrice maps regional price data.
func (m DetailsMapper) ToPrice(gameID int64, appID uint32, region domain.Region, data storefront.AppDetailsData, collectedAt time.Time) domain.GamePrice {
	price := domain.GamePrice{
		GameID:      gameID,
		AppID:       appID,
		Region:      region,
		IsFree:      data.IsFree,
		PriceState:  domain.PriceStateUnknown,
		CollectedAt: collectedAt,
	}
	if data.IsFree {
		price.PriceState = domain.PriceStateFree
		return price
	}
	if data.PriceOverview == nil {
		price.PriceState = domain.PriceStateUnpriced
		return price
	}
	price.PriceState = domain.PriceStatePriced
	price.Currency = data.PriceOverview.Currency
	price.Initial = int64(data.PriceOverview.Initial)
	price.Final = int64(data.PriceOverview.Final)
	price.DiscountPercent = int64(data.PriceOverview.DiscountPercent)
	price.InitialFormatted = data.PriceOverview.InitialFormatted
	price.FinalFormatted = data.PriceOverview.FinalFormatted
	return price
}

// ToMedia maps Store media.
func (m DetailsMapper) ToMedia(gameID int64, appID uint32, data storefront.AppDetailsData, collectedAt time.Time) domain.GameMedia {
	screenshots := make([]domain.Screenshot, 0, len(data.Screenshots))
	for _, screenshot := range data.Screenshots {
		screenshots = append(screenshots, domain.Screenshot{
			ID:           screenshot.ID,
			ThumbnailURL: screenshot.PathThumbnail,
			FullURL:      screenshot.PathFull,
		})
	}

	movies := make([]domain.Movie, 0, len(data.Movies))
	for _, movie := range data.Movies {
		movies = append(movies, domain.Movie{
			ID:           movie.ID,
			Name:         movie.Name,
			ThumbnailURL: movie.Thumbnail,
			WebM480URL:   movie.WebM.P480,
			WebMMaxURL:   movie.WebM.Max,
			MP4480URL:    movie.MP4.P480,
			MP4MaxURL:    movie.MP4.Max,
			DASHAV1URL:   movie.DASHAV1,
			DASHH264URL:  movie.DASHH264,
			HLSH264URL:   movie.HLSH264,
			Highlight:    movie.Highlight,
		})
	}

	return domain.GameMedia{
		GameID:           gameID,
		AppID:            appID,
		HeaderURL:        data.HeaderImage,
		CapsuleURL:       data.CapsuleImage,
		CapsuleV5URL:     data.CapsuleImageV5,
		BackgroundURL:    data.Background,
		BackgroundRawURL: data.BackgroundRaw,
		Screenshots:      screenshots,
		Movies:           movies,
		CollectedAt:      collectedAt,
	}
}

// ToRequirements maps system requirements.
func (m DetailsMapper) ToRequirements(gameID int64, appID uint32, data storefront.AppDetailsData, collectedAt time.Time) domain.SystemRequirements {
	return domain.SystemRequirements{
		GameID:      gameID,
		AppID:       appID,
		PC:          mapRequirements(data.PCRequirements),
		Mac:         mapRequirements(data.MacRequirements),
		Linux:       mapRequirements(data.LinuxRequirements),
		CollectedAt: collectedAt,
	}
}

func mapReleaseDate(value *storefront.StoreReleaseDate) domain.RawReleaseDate {
	if value == nil {
		return domain.RawReleaseDate{}
	}
	return domain.RawReleaseDate{ComingSoon: value.ComingSoon, DateText: value.Date}
}

func cloneDate(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	date := time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
	return &date
}

func positiveInt(value int) *int {
	if value <= 0 {
		return nil
	}
	return &value
}

func nonEmptyString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func mapPlatforms(value storefront.StorePlatforms) domain.PlatformSupport {
	return domain.PlatformSupport{Windows: value.Windows, Mac: value.Mac, Linux: value.Linux}
}

func mapSupportInfo(value *storefront.StoreSupportInfo) domain.SupportInfo {
	if value == nil {
		return domain.SupportInfo{}
	}
	return domain.SupportInfo{URL: value.URL, Email: value.Email}
}

func mapContentDescriptors(value *storefront.StoreContentDescriptors) domain.ContentDescriptors {
	if value == nil {
		return domain.ContentDescriptors{}
	}
	return domain.ContentDescriptors{IDs: append([]int(nil), value.IDs...), Notes: value.Notes}
}

func mapRequirements(value *storefront.StoreRequirements) domain.Requirements {
	if value == nil {
		return domain.Requirements{}
	}
	return domain.Requirements{Minimum: value.Minimum, Recommended: value.Recommended}
}

func mapRatings(data storefront.AppDetailsData) ([]domain.Rating, error) {
	if len(data.Ratings) == 0 {
		return nil, nil
	}
	var raw map[string]struct {
		Rating      string `json:"rating"`
		RequiredAge string `json:"required_age"`
	}
	if err := json.Unmarshal(data.Ratings, &raw); err != nil {
		return nil, fmt.Errorf("decode ratings: %w", err)
	}
	ratings := make([]domain.Rating, 0, len(raw))
	for board, rating := range raw {
		ratings = append(ratings, domain.Rating{
			Board:       board,
			Rating:      rating.Rating,
			RequiredAge: rating.RequiredAge,
		})
	}
	return ratings, nil
}
