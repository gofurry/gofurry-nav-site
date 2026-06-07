package steam

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/steam-go/web/storefront"
)

func TestNewsMapperFromPartnerEventCleansContentAndBuildsFallbackURL(t *testing.T) {
	t.Parallel()

	event := storefront.PartnerEvent{
		GID:               "event-1",
		AppID:             550,
		CommentCount:      7,
		ForumTopicID:      "forum-1",
		RTimeLastModified: 1700000100,
		AnnouncementBody: storefront.PartnerAnnouncementBody{
			GID:         "announce-1",
			EventGID:    "event-1",
			Headline:    "Patch Notes",
			Body:        `[b]Hello[/b] <script>alert(1)</script> world`,
			PostTime:    1700000000,
			Tags:        []string{"patch"},
			VoteUpCount: 12,
		},
		Raw: json.RawMessage(`{"gid":"event-1"}`),
	}

	got, err := NewNewsMapper().FromPartnerEvent(100, 550, domain.LanguageEN, event)
	if err != nil {
		t.Fatalf("FromPartnerEvent returned error: %v", err)
	}

	if got.URL != "https://store.steampowered.com/news/app/550/view/announce-1" {
		t.Fatalf("unexpected fallback URL: %q", got.URL)
	}
	if !strings.Contains(got.HTML, "<strong>Hello</strong>") {
		t.Fatalf("expected sanitized bold HTML, got %q", got.HTML)
	}
	if strings.Contains(got.HTML, "script") {
		t.Fatalf("expected script to be sanitized, got %q", got.HTML)
	}
	if got.PlainText == "" || got.Summary == "" {
		t.Fatalf("expected plain text and summary, got %#v", got)
	}
	if got.CommentCount != 7 || got.ForumTopicID != "forum-1" {
		t.Fatalf("expected event-level fallbacks, got %#v", got)
	}
	if len(got.RawEvent) == 0 {
		t.Fatal("expected raw event to be preserved")
	}
}
