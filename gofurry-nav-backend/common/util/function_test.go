package util

import "testing"

func TestTrustedProxyCIDRMatching(t *testing.T) {
	if !isTrustedProxyIP("127.0.0.1", "loopback") {
		t.Fatal("loopback should be trusted")
	}
	if !isTrustedProxyIP("10.6.0.11", "10.6.0.0/24") {
		t.Fatal("CIDR proxy should be trusted")
	}
	if isTrustedProxyIP("203.0.113.10", "10.6.0.0/24") {
		t.Fatal("public IP should not be trusted")
	}
}

func TestParseTokenRejectsInvalidTokenWithoutPanic(t *testing.T) {
	if claims, err := ParseToken("not-a-token"); err == nil || claims != nil {
		t.Fatalf("ParseToken() claims=%+v err=%v", claims, err)
	}
}
