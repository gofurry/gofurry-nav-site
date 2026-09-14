package main

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-site/tools/internal/schema"
)

// This acceptance check is read-only and can use the shared development GFN.
// Fresh database / rollback coverage remains in TestPostgresFreshAndBaselineAdoption.
func TestManagedAssetsDevSchema(t *testing.T) {
	dsn := os.Getenv("GOFURRY_GFN_SCHEMA_URL")
	if dsn == "" {
		t.Skip("set GOFURRY_GFN_SCHEMA_URL for the migrated dev GFN")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal("invalid database configuration")
	}
	defer db.Close()
	actual, err := schema.Inspect(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("GOFURRY_UPDATE_SCHEMA_SNAPSHOTS") == "1" {
		writeFinalExpectedSnapshot(t, root, "gfn", actual)
		return
	}
	expected, err := loadFinalExpected(root, "gfn")
	if err != nil {
		t.Fatal(err)
	}
	if difference := schema.Difference(expected, actual); difference != "" {
		t.Fatal(difference)
	}
}
