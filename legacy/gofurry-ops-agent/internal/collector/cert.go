package collector

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/model"
)

func collectCert(ctx context.Context, cfg config.CertCheckConfig) model.CertCheckResult {
	result := model.CertCheckResult{Name: cfg.Name, Host: cfg.Host}
	dialer := &net.Dialer{Timeout: cfg.Timeout.Duration}
	tlsCfg := &tls.Config{ServerName: cfg.Server, MinVersion: tls.VersionTLS12}
	checkCtx, cancel := context.WithTimeout(ctx, cfg.Timeout.Duration)
	defer cancel()

	tlsDialer := tls.Dialer{NetDialer: dialer, Config: tlsCfg}
	rawConn, err := tlsDialer.DialContext(checkCtx, "tcp", net.JoinHostPort(cfg.Host, cfg.Port))
	if err != nil {
		result.Status = statusFromError(err)
		result.ErrorMessage = err.Error()
		return result
	}
	defer rawConn.Close()
	conn, ok := rawConn.(*tls.Conn)
	if !ok {
		result.Status = "down"
		result.ErrorMessage = "unexpected tls connection type"
		return result
	}
	select {
	case <-checkCtx.Done():
		result.Status = "timeout"
		result.ErrorMessage = checkCtx.Err().Error()
		return result
	default:
	}
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		result.Status = "down"
		result.ErrorMessage = "no peer certificate"
		return result
	}
	cert := state.PeerCertificates[0]
	result.ExpiresAt = cert.NotAfter
	result.DaysRemaining = int(time.Until(cert.NotAfter).Hours() / 24)
	result.MatchedName = cert.VerifyHostname(cfg.Server) == nil
	result.Status = "ok"
	if result.DaysRemaining < 0 {
		result.Status = "down"
		result.ErrorMessage = "certificate expired"
	} else if result.DaysRemaining < cfg.WarnDays || !result.MatchedName {
		result.Status = "warn"
		if !result.MatchedName {
			result.ErrorMessage = "certificate hostname mismatch"
		}
	}
	return result
}
