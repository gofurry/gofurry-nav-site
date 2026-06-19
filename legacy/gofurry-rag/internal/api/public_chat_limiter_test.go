package api

import (
	"testing"
	"time"
)

func TestPublicChatLimiterRejectsNewKeysWhenFull(t *testing.T) {
	limiter := newPublicChatLimiter(10, time.Minute)
	limiter.maxKeys = 2

	if !limiter.Allow("198.51.100.1") || !limiter.Allow("198.51.100.2") {
		t.Fatal("expected first two keys to be allowed")
	}
	if limiter.Allow("198.51.100.3") {
		t.Fatal("expected third key to be rejected after cap")
	}
}

func TestPublicChatLimiterCleansExpiredKeys(t *testing.T) {
	limiter := newPublicChatLimiter(10, time.Minute)
	limiter.maxKeys = 1
	limiter.requests["198.51.100.1"] = publicChatWindow{resetAt: time.Now().Add(-time.Second), count: 1}

	if !limiter.Allow("198.51.100.2") {
		t.Fatal("expected new key after expired cleanup")
	}
	if _, ok := limiter.requests["198.51.100.1"]; ok {
		t.Fatal("expired key was not cleaned")
	}
}
