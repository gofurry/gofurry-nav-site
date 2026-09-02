package auditadmin

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRedactSnapshotRemovesSecretsRecursively(t *testing.T) {
	raw := `{"display_name":"Owner","password":"plain","nested":{"password_hash":"hash","token":"jwt","safe":"ok"}}`
	redacted := redactSnapshot(&raw)
	for _, forbidden := range []string{"plain", `:"hash"`, `:"jwt"`} {
		if strings.Contains(redacted, forbidden) {
			t.Fatalf("redacted snapshot still contains %q: %s", forbidden, redacted)
		}
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(redacted), &decoded); err != nil {
		t.Fatal(err)
	}
	nested := decoded["nested"].(map[string]any)
	if decoded["password"] != "[REDACTED]" || nested["password_hash"] != "[REDACTED]" || nested["token"] != "[REDACTED]" || nested["safe"] != "ok" {
		t.Fatalf("redacted snapshot lost safe content: %s", redacted)
	}
}

func TestRedactSnapshotRejectsInvalidJSON(t *testing.T) {
	raw := "postgres://user:password@host/database"
	if got := redactSnapshot(&raw); got != "{}" {
		t.Fatalf("invalid snapshot=%q, want empty object", got)
	}
}
