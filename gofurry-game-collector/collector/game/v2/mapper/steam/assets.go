package steam

import (
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/steam-go/addons/assets"
	"github.com/gofurry/steam-go/web/storefront"
)

const (
	assetSourceSteamStoreBrowse = assets.SourceStoreBrowse
	assetSourceSteamStorefront  = assets.SourceStorefrontAppDetails
)

// ToStoreBrowseAssets maps official StoreBrowse asset URLs into the unified asset model.
// Newer apps often use hashed paths, while older apps can still return legacy direct paths.
func (m DetailsMapper) ToStoreBrowseAssets(gameID int64, appID uint32, lang domain.Language, items []assets.URLItem, collectedAt time.Time) []domain.GameMediaAsset {
	return m.urlItemsToAssets(gameID, appID, string(lang), assetSourceSteamStoreBrowse, items, collectedAt)
}

// ToStorefrontAssets maps all Steam appdetails media URLs into the unified asset model.
func (m DetailsMapper) ToStorefrontAssets(gameID int64, appID uint32, data storefront.AppDetailsData, collectedAt time.Time) []domain.GameMediaAsset {
	out := make([]domain.GameMediaAsset, 0, len(data.Screenshots)*2+len(data.Movies)*8)

	for index, screenshot := range data.Screenshots {
		key := strconv.Itoa(screenshot.ID)
		out = appendAssetURL(out, storefrontAsset(gameID, appID, string(assets.KindScreenshotThumbnail), "screenshot", key, "", screenshot.PathThumbnail, "", index, nil, collectedAt))
		out = appendAssetURL(out, storefrontAsset(gameID, appID, string(assets.KindScreenshotFull), "screenshot", key, "", screenshot.PathFull, screenshot.PathThumbnail, index, nil, collectedAt))
	}

	for index, movie := range data.Movies {
		key := strconv.Itoa(movie.ID)
		extra := map[string]any{"highlight": movie.Highlight}
		out = appendAssetURL(out, storefrontAsset(gameID, appID, string(assets.KindMovieThumbnail), "movie", key, movie.Name, movie.Thumbnail, "", index, extra, collectedAt))
		out = appendAssetURL(out, storefrontAsset(gameID, appID, string(assets.KindMovieWebM480), "movie", key, movie.Name, movie.WebM.P480, movie.Thumbnail, index, extra, collectedAt))
		out = appendAssetURL(out, storefrontAsset(gameID, appID, string(assets.KindMovieWebMMax), "movie", key, movie.Name, movie.WebM.Max, movie.Thumbnail, index, extra, collectedAt))
		out = appendAssetURL(out, storefrontAsset(gameID, appID, string(assets.KindMovieMP4480), "movie", key, movie.Name, movie.MP4.P480, movie.Thumbnail, index, extra, collectedAt))
		out = appendAssetURL(out, storefrontAsset(gameID, appID, string(assets.KindMovieMP4Max), "movie", key, movie.Name, movie.MP4.Max, movie.Thumbnail, index, extra, collectedAt))
		out = appendAssetURL(out, storefrontAsset(gameID, appID, string(assets.KindMovieDASHAV1), "movie", key, movie.Name, movie.DASHAV1, movie.Thumbnail, index, extra, collectedAt))
		out = appendAssetURL(out, storefrontAsset(gameID, appID, string(assets.KindMovieDASHH264), "movie", key, movie.Name, movie.DASHH264, movie.Thumbnail, index, extra, collectedAt))
		out = appendAssetURL(out, storefrontAsset(gameID, appID, string(assets.KindMovieHLSH264), "movie", key, movie.Name, movie.HLSH264, movie.Thumbnail, index, extra, collectedAt))
	}

	return out
}

func (m DetailsMapper) urlItemsToAssets(gameID int64, appID uint32, lang string, source string, items []assets.URLItem, collectedAt time.Time) []domain.GameMediaAsset {
	out := make([]domain.GameMediaAsset, 0, len(items))
	for index, item := range items {
		rawURL := cleanAssetURL(item.URL)
		if rawURL == "" {
			continue
		}
		exists := true
		extra := storeBrowseExtra(item)
		out = append(out, domain.GameMediaAsset{
			GameID:      gameID,
			AppID:       appID,
			AssetType:   string(item.Kind),
			AssetFamily: assetFamilyForKind(string(item.Kind)),
			Source:      source,
			Language:    lang,
			MediaKey:    staticMediaKey(string(item.Kind)),
			Title:       item.Name,
			URL:         rawURL,
			Format:      assetFormat(string(item.Kind), rawURL),
			Exists:      &exists,
			Extra:       extra,
			SortOrder:   index,
			CheckedAt:   &collectedAt,
			CollectedAt: collectedAt,
		})
	}
	return out
}

func storefrontAsset(gameID int64, appID uint32, assetType string, family string, key string, title string, url string, thumbnailURL string, sortOrder int, extra any, collectedAt time.Time) domain.GameMediaAsset {
	exists := true
	cleanURL := cleanAssetURL(url)
	cleanThumbnailURL := cleanAssetURL(thumbnailURL)
	return domain.GameMediaAsset{
		GameID:       gameID,
		AppID:        appID,
		AssetType:    assetType,
		AssetFamily:  family,
		Source:       assetSourceSteamStorefront,
		MediaKey:     key,
		Title:        title,
		URL:          cleanURL,
		ThumbnailURL: cleanThumbnailURL,
		Format:       assetFormat(assetType, cleanURL),
		Exists:       &exists,
		Extra:        extra,
		SortOrder:    sortOrder,
		CheckedAt:    &collectedAt,
		CollectedAt:  collectedAt,
	}
}

func appendAssetURL(items []domain.GameMediaAsset, item domain.GameMediaAsset) []domain.GameMediaAsset {
	if strings.TrimSpace(item.URL) == "" {
		return items
	}
	return append(items, item)
}

func storeBrowseExtra(item assets.URLItem) map[string]any {
	extra := map[string]any{}
	if value := strings.TrimSpace(item.Digest); value != "" {
		extra["digest"] = value
	}
	if value := strings.TrimSpace(item.Filename); value != "" {
		extra["filename"] = value
	}
	if value := strings.TrimSpace(item.Source); value != "" {
		extra["source"] = value
	}
	return extra
}

func cleanAssetURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	return parsed.String()
}

func staticMediaKey(kind string) string {
	if kind == "" {
		return "default"
	}
	return kind
}

func assetFamilyForKind(kind string) string {
	switch {
	case strings.HasPrefix(kind, "library_"):
		return "library"
	case strings.HasPrefix(kind, "screenshot_"):
		return "screenshot"
	case strings.HasPrefix(kind, "movie_"):
		return "movie"
	case strings.Contains(kind, "background"):
		return "background"
	case strings.Contains(kind, "icon") || strings.Contains(kind, "logo"):
		return "icon"
	default:
		return "store"
	}
}

func assetFormat(kind string, rawURL string) string {
	rawURL = cleanAssetURL(rawURL)
	switch {
	case strings.Contains(kind, "dash") || strings.Contains(kind, "hls"):
		return "playlist"
	case strings.Contains(kind, "webm") || strings.Contains(kind, "mp4"):
		return "video"
	}
	switch strings.ToLower(filepath.Ext(rawURL)) {
	case ".webm", ".mp4":
		return "video"
	case ".mpd", ".m3u8":
		return "playlist"
	case ".png", ".jpg", ".jpeg", ".webp", ".gif":
		return "image"
	default:
		return ""
	}
}
