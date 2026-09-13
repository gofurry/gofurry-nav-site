// Package assets owns bounded immutable object writes. Business references and
// audit transactions remain in Nav Admin; no replication database is required.
package assets

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

const CacheControl = "public, max-age=31536000, immutable"
const ProbeKey = "system/probes/cdn.bin"
const MaxSize = 5 << 20

type Object struct {
	Key         string `json:"object_key"`
	Data        []byte `json:"-"`
	ContentType string `json:"content_type"`
	SHA256      string `json:"sha256"`
	Kind        string `json:"asset_kind"`
}
type Info struct {
	Exists       bool   `json:"exists"`
	Size         int64  `json:"size"`
	ContentType  string `json:"content_type"`
	SHA256       string `json:"sha256"`
	Kind         string `json:"asset_kind"`
	CacheControl string `json:"cache_control"`
}
type ObjectStore interface {
	Put(context.Context, Object) error
	Head(context.Context, string) (Info, error)
	Get(context.Context, string) ([]byte, error)
}
type Service struct{ Primary, Mirror ObjectStore }
type Publication struct {
	ObjectKey string   `json:"object_key"`
	Primary   string   `json:"primary"`
	Mirror    string   `json:"mirror"`
	Warnings  []string `json:"warnings"`
}

func NewObject(kind string, siteID int64, filename string, data []byte) (Object, error) {
	o := Object{Data: data, Kind: kind}
	contentType, ext, err := Validate(kind, filename, data)
	if err != nil {
		return o, err
	}
	hash := sha256.Sum256(data)
	o.SHA256 = hex.EncodeToString(hash[:])
	o.ContentType = contentType
	switch kind {
	case "site-icon":
		if siteID <= 0 {
			return o, errors.New("invalid site id")
		}
		o.Key = fmt.Sprintf("nav/sites/%d/icon/%s%s", siteID, o.SHA256[:32], ext)
	case "hero-desktop", "hero-mobile":
		o.Key = "nav/hero/" + strings.TrimPrefix(kind, "hero-") + "/" + o.SHA256[:32] + ".avif"
	case "pattern":
		o.Key = "nav/patterns/" + o.SHA256[:32] + ".svg"
	default:
		return o, errors.New("unsupported asset kind")
	}
	return o, nil
}

func (s *Service) Publish(ctx context.Context, o Object) (Publication, error) {
	p := Publication{ObjectKey: o.Key, Primary: "failed", Mirror: "failed", Warnings: []string{}}
	if s == nil || s.Primary == nil {
		return p, errors.New("COS primary is not configured")
	}
	if err := ensure(ctx, s.Primary, o); err != nil {
		return p, fmt.Errorf("COS primary publication failed: %w", err)
	}
	p.Primary = "ready"
	if s.Mirror != nil {
		for i := 0; i < 2; i++ {
			if err := ensure(ctx, s.Mirror, o); err == nil {
				p.Mirror = "ready"
				return p, nil
			}
			if i == 0 {
				select {
				case <-ctx.Done():
					p.Warnings = append(p.Warnings, "R2 mirror sync failed")
					return p, nil
				case <-time.After(200 * time.Millisecond):
				}
			}
		}
	}
	p.Warnings = append(p.Warnings, "R2 mirror sync failed")
	return p, nil
}

func ensure(ctx context.Context, store ObjectStore, o Object) error {
	head, err := store.Head(ctx, o.Key)
	if err != nil {
		return err
	}
	if head.Exists {
		return verify(head, o)
	}
	if err := store.Put(ctx, o); err != nil {
		return err
	}
	head, err = store.Head(ctx, o.Key)
	if err != nil {
		return err
	}
	return verify(head, o)
}
func verify(info Info, o Object) error {
	if !info.Exists || info.Size != int64(len(o.Data)) || info.SHA256 != o.SHA256 || info.ContentType != o.ContentType || info.Kind != o.Kind || info.CacheControl != CacheControl {
		return errors.New("immutable object metadata mismatch")
	}
	return nil
}

func (s *Service) RepairMirror(ctx context.Context, key string) error {
	if !ValidKey(key) || key == ProbeKey {
		return errors.New("invalid managed object key")
	}
	if s == nil || s.Primary == nil || s.Mirror == nil {
		return errors.New("both object stores must be configured")
	}
	info, err := s.Primary.Head(ctx, key)
	if err != nil {
		return err
	}
	if !info.Exists {
		return errors.New("primary object not found")
	}
	data, err := s.Primary.Get(ctx, key)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(data)
	if hex.EncodeToString(hash[:]) != info.SHA256 || !strings.Contains(key, info.SHA256[:32]) {
		return errors.New("primary object hash mismatch")
	}
	o := Object{Key: key, Data: data, SHA256: info.SHA256, Kind: info.Kind, ContentType: info.ContentType}
	if err := verify(info, o); err != nil {
		return err
	}
	// Repair may replace corrupt mirror bytes, always using the immutable primary.
	if err := s.Mirror.Put(ctx, o); err != nil {
		return err
	}
	head, err := s.Mirror.Head(ctx, key)
	if err != nil {
		return err
	}
	return verify(head, o)
}
