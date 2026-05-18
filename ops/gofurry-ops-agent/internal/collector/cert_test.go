package collector

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/config"
)

func TestCollectCertHonorsCanceledContext(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	result := collectCert(ctx, config.CertCheckConfig{
		Name:     "local",
		Host:     "127.0.0.1",
		Port:     strconv.Itoa(listener.Addr().(*net.TCPAddr).Port),
		Server:   "localhost",
		Timeout:  config.Duration{Duration: 5 * time.Second},
		WarnDays: 30,
	})
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("expected canceled cert check to return quickly, took %s", elapsed)
	}
	if result.Status != "timeout" {
		t.Fatalf("expected timeout status for canceled context, got %#v", result)
	}
}
