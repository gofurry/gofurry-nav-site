package security

import "testing"

func TestSignAndVerify(t *testing.T) {
	body := []byte(`{"node_id":"cn-business-a"}`)
	sig := Sign("secret", "2026-05-18T12:00:00Z", "cn-business-a", body)
	if sig == "" || sig[:7] != SignaturePrefix {
		t.Fatalf("unexpected signature: %s", sig)
	}
	if !Verify("secret", "2026-05-18T12:00:00Z", "cn-business-a", body, sig) {
		t.Fatal("expected valid signature")
	}
	if Verify("other", "2026-05-18T12:00:00Z", "cn-business-a", body, sig) {
		t.Fatal("expected invalid token")
	}
	if Verify("secret", "2026-05-18T12:00:00Z", "other", body, sig) {
		t.Fatal("expected invalid node")
	}
}
