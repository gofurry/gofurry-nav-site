package reporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/model"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/security"
)

type Reporter struct {
	cfg    config.CenterConfig
	nodeID string
	client *http.Client
}

func New(cfg config.CenterConfig, nodeID string) *Reporter {
	return &Reporter{
		cfg:    cfg,
		nodeID: nodeID,
		client: &http.Client{Timeout: cfg.Timeout.Duration},
	}
}

func (r *Reporter) Send(ctx context.Context, payload model.Payload) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return body, r.SendRaw(ctx, body)
}

func (r *Reporter) SendRaw(ctx context.Context, body []byte) error {
	var lastErr error
	attempts := r.cfg.Retries
	if attempts <= 0 {
		attempts = 1
	}
	for i := 0; i < attempts; i++ {
		if err := r.sendRawOnce(ctx, body); err != nil {
			lastErr = err
			if i+1 < attempts {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Duration(i+1) * 300 * time.Millisecond):
				}
			}
			continue
		}
		return nil
	}
	return lastErr
}

func (r *Reporter) sendRawOnce(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	timestamp := time.Now().UTC().Format(time.RFC3339)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.cfg.Token)
	req.Header.Set("X-GoFurry-Node-ID", r.nodeID)
	req.Header.Set("X-GoFurry-Timestamp", timestamp)
	req.Header.Set("X-GoFurry-Signature", security.Sign(r.cfg.Token, timestamp, r.nodeID, body))

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("center returned %s: %s", resp.Status, strings.TrimSpace(string(msg)))
	}
	return nil
}
