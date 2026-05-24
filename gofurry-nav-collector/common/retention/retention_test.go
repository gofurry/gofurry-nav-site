package retention

import (
	"strings"
	"testing"
)

func TestBuildDeleteSQLUsesStableRetentionShape(t *testing.T) {
	sql, err := BuildDeleteSQL("gfn_collector_log_ping")
	if err != nil {
		t.Fatalf("BuildDeleteSQL() error = %v", err)
	}

	for _, want := range []string{
		"FROM gfn_collector_log_ping",
		"DELETE FROM gfn_collector_log_ping target",
		"PARTITION BY name",
		"ORDER BY create_time DESC, id DESC",
		"WHERE ranked.rn > ?",
		"LIMIT ?",
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("BuildDeleteSQL() missing %q in:\n%s", want, sql)
		}
	}
}

func TestBuildDeleteSQLDoesNotUseLegacyLimitedWindow(t *testing.T) {
	sql, err := BuildDeleteSQL("gfn_collector_log_http")
	if err != nil {
		t.Fatalf("BuildDeleteSQL() error = %v", err)
	}

	if strings.Contains(sql, "FROM gfn_collector_log_http\n\t\t\t\tLIMIT") {
		t.Fatalf("BuildDeleteSQL() should not limit the ranked source window:\n%s", sql)
	}
}

func TestBuildObservationDeleteSQLPartitionsBySiteAndProtocol(t *testing.T) {
	sql, err := BuildObservationDeleteSQL("gfn_collector_observation")
	if err != nil {
		t.Fatalf("BuildObservationDeleteSQL() error = %v", err)
	}

	for _, want := range []string{
		"PARTITION BY site_id, protocol",
		"ORDER BY observed_at DESC, id DESC",
		"FROM gfn_collector_observation",
		"WHERE protocol = ?",
		"LIMIT ?",
		"DELETE FROM gfn_collector_observation target",
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("BuildObservationDeleteSQL() missing %q in:\n%s", want, sql)
		}
	}
}

func TestBuildDeleteSQLRejectsUnknownTable(t *testing.T) {
	if _, err := BuildDeleteSQL("gfn_site"); err == nil {
		t.Fatal("BuildDeleteSQL() should reject table names outside the allowlist")
	}
}

func TestBuildObservationDeleteSQLRejectsUnknownTable(t *testing.T) {
	if _, err := BuildObservationDeleteSQL("gfn_site"); err == nil {
		t.Fatal("BuildObservationDeleteSQL() should reject table names outside the allowlist")
	}
}
