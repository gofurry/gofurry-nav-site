package steam

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/steam-go/addons/markup"
	"github.com/gofurry/steam-go/web/storefront"
)

const defaultSummaryRunes = 180

// NewsMapper converts steam-go Store events into collector v2 news domain models.
type NewsMapper struct {
	SummaryRunes int
}

// NewNewsMapper returns a mapper with conservative defaults.
func NewNewsMapper() NewsMapper {
	return NewsMapper{SummaryRunes: defaultSummaryRunes}
}

// FromPartnerEvent maps one Store event into one v2 news item.
func (m NewsMapper) FromPartnerEvent(gameID int64, appID uint32, lang domain.Language, event storefront.PartnerEvent) (domain.GameNews, error) {
	body := event.AnnouncementBody
	rawBody := body.Body

	html, err := markup.CleanSteamContent(rawBody)
	if err != nil {
		return domain.GameNews{}, fmt.Errorf("clean steam content: %w", err)
	}
	plainText, err := markup.PlainText(rawBody)
	if err != nil {
		return domain.GameNews{}, fmt.Errorf("plain steam content: %w", err)
	}
	summary, err := markup.Summary(rawBody, m.summaryRunes())
	if err != nil {
		return domain.GameNews{}, fmt.Errorf("summarize steam content: %w", err)
	}

	eventGID := firstNonEmpty(body.EventGID, event.GID)
	announcementGID := firstNonEmpty(body.GID, eventGID)
	url := body.URL
	if url == "" {
		url = domain.SteamNewsURL(appID, announcementGID)
	}

	rawEvent := event.Raw
	if len(rawEvent) == 0 {
		rawEvent, err = json.Marshal(event)
		if err != nil {
			return domain.GameNews{}, fmt.Errorf("marshal raw event fallback: %w", err)
		}
	}

	return domain.GameNews{
		GameID:          gameID,
		AppID:           appID,
		Language:        lang,
		EventGID:        eventGID,
		AnnouncementGID: announcementGID,
		ForumTopicID:    firstNonEmpty(body.ForumTopicID, event.ForumTopicID),
		Headline:        body.Headline,
		RawBody:         rawBody,
		HTML:            html,
		PlainText:       plainText,
		Summary:         summary,
		URL:             url,
		Tags:            append([]string(nil), body.Tags...),
		VoteUpCount:     body.VoteUpCount,
		VoteDownCount:   body.VoteDownCount,
		CommentCount:    firstNonZero(body.CommentCount, event.CommentCount),
		RawEvent:        rawEvent,
		PublishedAt:     unixTime(body.PostTime),
		UpdatedAt:       unixTime(firstNonZero64(body.UpdateTime, event.RTimeLastModified)),
		CollectedAt:     time.Now(),
	}, nil
}

func (m NewsMapper) summaryRunes() int {
	if m.SummaryRunes <= 0 {
		return defaultSummaryRunes
	}
	return m.SummaryRunes
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func firstNonZero64(values ...int64) int64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func unixTime(value int64) time.Time {
	if value <= 0 {
		return time.Time{}
	}
	return time.Unix(value, 0)
}
