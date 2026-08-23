// Command db-baseline adopts an exact pre-Goose GoFurry database by recording
// only the audited baseline version. It never executes business DDL.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	databaseLabel := flag.String("database", "", "database contract to adopt: gfg, gfn, or gfa")
	baselineVersion := flag.Int64("baseline-version", 0, "exact audited baseline version")
	confirm := flag.Bool("confirm-adopt", false, "confirm the one-time version-store write")
	flag.Parse()

	if !*confirm {
		exitf("refusing adoption without -confirm-adopt")
	}
	dsn := os.Getenv("GOFURRY_DATABASE_URL")
	if dsn == "" {
		exitf("GOFURRY_DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		exitf("open PostgreSQL connection: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		exitf("connect to PostgreSQL: %v", err)
	}
	if err := adopt(ctx, db, adoptOptions{
		DatabaseLabel:   *databaseLabel,
		BaselineVersion: *baselineVersion,
		RequireDBName:   true,
	}); err != nil {
		exitf("baseline adoption failed: %v", err)
	}
	fmt.Printf("adopted %s baseline version %d; no business DDL was executed\n", *databaseLabel, *baselineVersion)
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
