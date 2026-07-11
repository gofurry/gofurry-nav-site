package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
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
	signature = strings.TrimSpace(signature)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1
}
