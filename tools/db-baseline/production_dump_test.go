package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-site/tools/internal/schema"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// TestProductionSchemaDumpAdoption characterizes an operator-supplied, schema-only
// pg_dump without ever connecting to production. It is skipped unless both dump
// paths are supplied explicitly.
func TestProductionSchemaDumpAdoption(t *testing.T) {
	adminDSN := integrationAdminDSN(t)
	dumps := map[string]string{
		"gfg": os.Getenv("GOFURRY_GFG_SCHEMA_DUMP"),
		"gfn": os.Getenv("GOFURRY_GFN_SCHEMA_DUMP"),
	}
	for label, path := range dumps {
		if path == "" {
			t.Skipf("set GOFURRY_GFG_SCHEMA_DUMP and GOFURRY_GFN_SCHEMA_DUMP to validate production schema dumps")
		}
		t.Run(label, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			adminDB := openDatabase(t, adminDSN, "postgres")
			defer adminDB.Close()
			databaseName := temporaryDatabaseName(label, "dump")
			createDatabase(t, ctx, adminDB, databaseName)
			defer dropDatabase(t, adminDB, databaseName)

			dump, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s schema dump: %v", label, err)
			}
			restoreDB := openSimpleProtocolDatabase(t, adminDSN, databaseName)
			if _, err := restoreDB.ExecContext(ctx, stripPSQLMetaCommands(string(dump))); err != nil {
				restoreDB.Close()
				t.Fatalf("restore %s schema dump into temporary database: %v", label, err)
			}
			if err := restoreDB.Close(); err != nil {
				t.Fatalf("close %s schema restore connection: %v", label, err)
			}
			db := openDatabase(t, adminDSN, databaseName)
			defer db.Close()

			expected, err := loadExpected(label)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := schema.Inspect(ctx, db)
			if err != nil {
				t.Fatalf("inspect restored %s schema: %v", label, err)
			}
			if differences := snapshotDifferences(expected, actual); len(differences) > 0 {
				t.Fatalf("restored %s schema has %d baseline difference(s):\n%s", label, len(differences), strings.Join(differences, "\n"))
			}
			fingerprint, err := schema.Fingerprint(actual)
			if err != nil {
				t.Fatalf("fingerprint restored %s schema: %v", label, err)
			}
			t.Logf("restored %s schema fingerprint: %s", label, fingerprint)

			if err := adopt(ctx, db, adoptOptions{DatabaseLabel: label, BaselineVersion: expectedBaselineVersion}); err != nil {
				t.Fatalf("adopt restored %s schema dump: %v", label, err)
			}
		})
	}
}

func openSimpleProtocolDatabase(t *testing.T, dsn, databaseName string) *sql.DB {
	t.Helper()
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.Database = databaseName
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	config.RuntimeParams["application_name"] = "gofurry-foundation-dump-characterization"
	return stdlib.OpenDB(*config)
}

func stripPSQLMetaCommands(dump string) string {
	dump = strings.ReplaceAll(dump, "\r\n", "\n")
	lines := strings.Split(dump, "\n")
	kept := lines[:0]
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), `\`) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

func snapshotDifferences(expected, actual schema.Snapshot) []string {
	expectedJSON, _ := json.Marshal(expected)
	actualJSON, _ := json.Marshal(actual)
	var expectedValue any
	var actualValue any
	_ = json.Unmarshal(expectedJSON, &expectedValue)
	_ = json.Unmarshal(actualJSON, &actualValue)
	var differences []string
	diffJSONValue("$", expectedValue, actualValue, &differences)
	return differences
}

func diffJSONValue(path string, expected, actual any, differences *[]string) {
	expectedMap, expectedIsMap := expected.(map[string]any)
	actualMap, actualIsMap := actual.(map[string]any)
	if expectedIsMap && actualIsMap {
		keys := make(map[string]struct{}, len(expectedMap)+len(actualMap))
		for key := range expectedMap {
			keys[key] = struct{}{}
		}
		for key := range actualMap {
			keys[key] = struct{}{}
		}
		orderedKeys := make([]string, 0, len(keys))
		for key := range keys {
			orderedKeys = append(orderedKeys, key)
		}
		sort.Strings(orderedKeys)
		for _, key := range orderedKeys {
			expectedChild, expectedOK := expectedMap[key]
			actualChild, actualOK := actualMap[key]
			if !expectedOK || !actualOK {
				*differences = append(*differences, formatDifference(path+"."+key, expectedChild, actualChild))
				continue
			}
			diffJSONValue(path+"."+key, expectedChild, actualChild, differences)
		}
		return
	}

	expectedSlice, expectedIsSlice := expected.([]any)
	actualSlice, actualIsSlice := actual.([]any)
	if expectedIsSlice && actualIsSlice {
		if expectedNamed, ok := keyedJSONValues(expectedSlice); ok {
			if actualNamed, actualOK := keyedJSONValues(actualSlice); actualOK {
				diffJSONValue(path, expectedNamed, actualNamed, differences)
				return
			}
		}
		maximum := len(expectedSlice)
		if len(actualSlice) > maximum {
			maximum = len(actualSlice)
		}
		for i := 0; i < maximum; i++ {
			if i >= len(expectedSlice) || i >= len(actualSlice) {
				var expectedChild, actualChild any
				if i < len(expectedSlice) {
					expectedChild = expectedSlice[i]
				}
				if i < len(actualSlice) {
					actualChild = actualSlice[i]
				}
				*differences = append(*differences, formatDifference(fmt.Sprintf("%s[%d]", path, i), expectedChild, actualChild))
				continue
			}
			diffJSONValue(fmt.Sprintf("%s[%d]", path, i), expectedSlice[i], actualSlice[i], differences)
		}
		return
	}

	if fmt.Sprint(expected) != fmt.Sprint(actual) {
		*differences = append(*differences, formatDifference(path, expected, actual))
	}
}

func keyedJSONValues(values []any) (map[string]any, bool) {
	if len(values) == 0 {
		return nil, false
	}
	for _, key := range []string{"name", "identity"} {
		keyed := make(map[string]any, len(values))
		valid := true
		for _, value := range values {
			object, ok := value.(map[string]any)
			if !ok {
				valid = false
				break
			}
			name, ok := object[key].(string)
			if !ok || name == "" {
				valid = false
				break
			}
			keyed[name] = value
		}
		if valid {
			return keyed, true
		}
	}
	return nil, false
}

func formatDifference(path string, expected, actual any) string {
	return fmt.Sprintf("%s: expected %s; actual %s", path, compactJSON(expected), compactJSON(actual))
}

func compactJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}
