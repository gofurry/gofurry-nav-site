package prizekey

import (
	"net/url"
	"strings"
	"testing"
)

func TestCacheKeys(t *testing.T) {
	emailKey := EmailLockKey(42, "user@example.com")
	if emailKey != "game:v2:prize:email:42:user@example.com" {
		t.Fatalf("unexpected email lock key: %s", emailKey)
	}

	participationKey := ParticipationKey(42, "token-1")
	if participationKey != "game:v2:prize:participation:42:token-1" {
		t.Fatalf("unexpected participation key: %s", participationKey)
	}
}

func TestActivationLink(t *testing.T) {
	link := ActivationLink(42, "token with space")
	if strings.Contains(link, "/api/v1/") {
		t.Fatalf("activation link must not point to v1: %s", link)
	}
	if !strings.Contains(link, "/api/v2/game/prizes/participation/activation") {
		t.Fatalf("activation link must point to v2 prize activation endpoint: %s", link)
	}

	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("invalid activation link: %v", err)
	}
	if got := u.Query().Get("id"); got != "42" {
		t.Fatalf("unexpected id query: %s", got)
	}
	if got := u.Query().Get("key"); got != "token with space" {
		t.Fatalf("unexpected key query: %s", got)
	}
}
