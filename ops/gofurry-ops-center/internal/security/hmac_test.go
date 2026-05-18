package security

import (
	"testing"
	"time"
)

func TestVerifySignature(t *testing.T) {
	body := []byte(`{"node_id":"cn"}`)
	sig := Sign("token", "2026-05-18T12:00:00Z", "cn", body)
	if !Verify("token", "2026-05-18T12:00:00Z", "cn", body, sig) {
		t.Fatal("expected valid signature")
	}
	if Verify("bad", "2026-05-18T12:00:00Z", "cn", body, sig) {
		t.Fatal("expected invalid signature")
	}
}

func TestSession(t *testing.T) {
	token, err := NewSession("secret", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !VerifySession("secret", token) {
		t.Fatal("expected valid session")
	}
	if VerifySession("other", token) {
		t.Fatal("expected invalid session")
	}
}

func TestSessionUsesRandomNonce(t *testing.T) {
	first, err := NewSession("secret", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewSession("secret", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("expected unique session tokens")
	}
}

func TestCheckTimestamp(t *testing.T) {
	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	if err := CheckTimestamp("2026-05-18T12:03:00Z", 5*time.Minute, now); err != nil {
		t.Fatal(err)
	}
	if err := CheckTimestamp("2026-05-18T11:00:00Z", 5*time.Minute, now); err == nil {
		t.Fatal("expected stale timestamp error")
	}
}
