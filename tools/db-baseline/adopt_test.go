package main

import (
	"strings"
	"testing"

	"github.com/gofurry/gofurry-nav-site/tools/internal/schema"
)

func TestExpectedSnapshotsLoad(t *testing.T) {
	t.Parallel()
	tests := []struct {
		database                     string
		tables, sequences, functions int
		extensions, nonUniqueIndexes int
	}{
		{database: "gfg", tables: 22, sequences: 6, functions: 1, extensions: 1, nonUniqueIndexes: 43},
		{database: "gfn", tables: 12, sequences: 0, functions: 0, extensions: 0, nonUniqueIndexes: 38},
		{database: "gfa", tables: 2, sequences: 2, functions: 0, extensions: 0, nonUniqueIndexes: 0},
	}
	for _, test := range tests {
		database := test.database
		t.Run(database, func(t *testing.T) {
			t.Parallel()
			snapshot, err := loadExpected(database)
			if err != nil {
				t.Fatal(err)
			}
			nonUniqueIndexes := 0
			for _, table := range snapshot.Tables {
				for _, index := range table.Indexes {
					if !strings.HasPrefix(index.Definition, "CREATE UNIQUE INDEX") {
						nonUniqueIndexes++
					}
				}
			}
			if len(snapshot.Tables) != test.tables || len(snapshot.Sequences) != test.sequences ||
				len(snapshot.Functions) != test.functions || len(snapshot.Extensions) != test.extensions ||
				nonUniqueIndexes != test.nonUniqueIndexes {
				t.Fatalf("unexpected audited inventory: tables=%d sequences=%d functions=%d extensions=%d indexes=%d",
					len(snapshot.Tables), len(snapshot.Sequences), len(snapshot.Functions), len(snapshot.Extensions), nonUniqueIndexes)
			}
		})
	}
}

func TestDifferenceReportsDrift(t *testing.T) {
	t.Parallel()
	expected := schema.Snapshot{FormatVersion: 1, Tables: []schema.Table{{Name: "one"}}}
	actual := schema.Snapshot{FormatVersion: 1, Tables: []schema.Table{{Name: "two"}}}
	if difference := schema.Difference(expected, actual); difference == "" {
		t.Fatal("expected schema difference")
	}
}
