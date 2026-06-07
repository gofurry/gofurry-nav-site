package steam

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/steam-go/web/storefront"
)

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
		GameID:             gameID,
		AppID:              appID,
		Type:               data.Type,
		Name:               data.Name,
		IsFree:             data.IsFree,
		Website:            data.Website,
		HeaderURL:          data.HeaderImage,
		Developers:         append([]string(nil), data.Developers...),
		Publishers:         append([]string(nil), data.Publishers...),
		Release:            mapReleaseDate(data.ReleaseDate),
		Platforms:          mapPlatforms(data.Platforms),
		SupportedLanguages: data.SupportedLanguages,
		SupportInfo:        mapSupportInfo(data.SupportInfo),
		ContentDescriptors: mapContentDescriptors(data.ContentDescriptors),
		Ratings:            ratings,
		CollectedAt:        collectedAt,
	}, nil
}

// ToLocalized maps language-specific copy.
func (m DetailsMapper) ToLocalized(gameID int64, appID uint32, lang domain.Language, data storefront.AppDetailsData, collectedAt time.Time) domain.GameLocalizedDetails {
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

// ToPrice maps regional price data.
func (m DetailsMapper) ToPrice(gameID int64, appID uint32, region domain.Region, data storefront.AppDetailsData, collectedAt time.Time) domain.GamePrice {
	price := domain.GamePrice{
		GameID:      gameID,
		AppID:       appID,
		Region:      region,
		IsFree:      data.IsFree,
		CollectedAt: collectedAt,
	}
	if data.IsFree {
		return price
	}
	if data.PriceOverview == nil {
		return price
	}
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

func mapReleaseDate(value *storefront.StoreReleaseDate) domain.ReleaseDate {
	if value == nil {
		return domain.ReleaseDate{}
	}
	return domain.ReleaseDate{ComingSoon: value.ComingSoon, DateText: value.Date}
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
