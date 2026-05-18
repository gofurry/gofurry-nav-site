package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const SignaturePrefix = "sha256="

func Sign(token, timestamp, nodeID string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(token))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("\n"))
	mac.Write([]byte(nodeID))
	mac.Write([]byte("\n"))
	mac.Write(body)
	return SignaturePrefix + hex.EncodeToString(mac.Sum(nil))
}

func Verify(token, timestamp, nodeID string, body []byte, signature string) bool {
	expected := Sign(token, timestamp, nodeID, body)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(strings.TrimSpace(signature))) == 1
}

func NewSession(secret string, ttl time.Duration) string {
	expires := time.Now().Add(ttl).Unix()
	payload := strconv.FormatInt(expires, 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(payload + ":" + sig))
}

func VerifySession(secret, token string) bool {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil {
		return false
	}
	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 {
		return false
	}
	expires, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || time.Now().Unix() > expires {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0]))
	expected := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(expected), []byte(parts[1])) == 1
}

func CheckTimestamp(value string, window time.Duration, now time.Time) error {
	ts, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return err
	}
	if ts.Before(now.Add(-window)) || ts.After(now.Add(window)) {
		return fmt.Errorf("timestamp outside allowed window")
	}
	return nil
}
