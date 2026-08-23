// Package schema captures the PostgreSQL application schema in a stable form.
// It intentionally ignores owners, ACLs, volatile sequence values, extension-
// owned objects, and Goose's own version table.
package schema

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

const GooseTable = "goose_db_version"

type Queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Snapshot struct {
	FormatVersion int         `json:"format_version"`
	Extensions    []Extension `json:"extensions"`
	Sequences     []Sequence  `json:"sequences"`
	Tables        []Table     `json:"tables"`
	Functions     []Function  `json:"functions"`
}

type Extension struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

type Sequence struct {
	Name        string `json:"name"`
	DataType    string `json:"data_type"`
	Start       int64  `json:"start"`
	Increment   int64  `json:"increment"`
	Minimum     int64  `json:"minimum"`
	Maximum     int64  `json:"maximum"`
	Cache       int64  `json:"cache"`
	Cycle       bool   `json:"cycle"`
	OwnedTable  string `json:"owned_table,omitempty"`
	OwnedColumn string `json:"owned_column,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

type Table struct {
	Name        string       `json:"name"`
	Comment     string       `json:"comment,omitempty"`
	Columns     []Column     `json:"columns"`
	Constraints []Constraint `json:"constraints"`
	Indexes     []Index      `json:"indexes"`
	Triggers    []Trigger    `json:"triggers"`
}

type Column struct {
	Name       string `json:"name"`
	DataType   string `json:"data_type"`
	NotNull    bool   `json:"not_null"`
	HasDefault bool   `json:"has_default"`
	Default    string `json:"default,omitempty"`
	Identity   string `json:"identity,omitempty"`
	Generated  string `json:"generated,omitempty"`
	Collation  string `json:"collation,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

type Constraint struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
	Comment    string `json:"comment,omitempty"`
}

type Index struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
	Comment    string `json:"comment,omitempty"`
}

type Trigger struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
	Comment    string `json:"comment,omitempty"`
}

type Function struct {
	Identity   string `json:"identity"`
	Definition string `json:"definition"`
	Comment    string `json:"comment,omitempty"`
}

func Inspect(ctx context.Context, db Queryer) (Snapshot, error) {
	snapshot := Snapshot{FormatVersion: 1}
	if err := loadExtensions(ctx, db, &snapshot); err != nil {
		return Snapshot{}, err
	}
	if err := loadSequences(ctx, db, &snapshot); err != nil {
		return Snapshot{}, err
	}
	if err := loadTables(ctx, db, &snapshot); err != nil {
		return Snapshot{}, err
	}
	if err := loadFunctions(ctx, db, &snapshot); err != nil {
		return Snapshot{}, err
	}
	return snapshot, nil
}

func Marshal(snapshot Snapshot) ([]byte, error) {
	return json.MarshalIndent(snapshot, "", "  ")
}

func Unmarshal(data []byte) (Snapshot, error) {
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, err
	}
	if snapshot.FormatVersion != 1 {
		return Snapshot{}, fmt.Errorf("unsupported schema snapshot format %d", snapshot.FormatVersion)
	}
	return snapshot, nil
}

func Fingerprint(snapshot Snapshot) (string, error) {
	data, err := Marshal(snapshot)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func Difference(expected, actual Snapshot) string {
	expectedJSON, _ := Marshal(expected)
	actualJSON, _ := Marshal(actual)
	if string(expectedJSON) == string(actualJSON) {
		return ""
	}
	expectedLines := strings.Split(string(expectedJSON), "\n")
	actualLines := strings.Split(string(actualJSON), "\n")
	limit := len(expectedLines)
	if len(actualLines) < limit {
		limit = len(actualLines)
	}
	for i := 0; i < limit; i++ {
		if expectedLines[i] != actualLines[i] {
			return fmt.Sprintf("first difference at snapshot line %d: expected %s; actual %s", i+1, strings.TrimSpace(expectedLines[i]), strings.TrimSpace(actualLines[i]))
		}
	}
	return fmt.Sprintf("snapshot length differs: expected %d lines; actual %d lines", len(expectedLines), len(actualLines))
}

func loadExtensions(ctx context.Context, db Queryer, snapshot *Snapshot) error {
	rows, err := db.QueryContext(ctx, `
		select e.extname, n.nspname
		from pg_catalog.pg_extension e
		join pg_catalog.pg_namespace n on n.oid = e.extnamespace
		where e.extname <> 'plpgsql'
		order by e.extname`)
	if err != nil {
		return fmt.Errorf("query extensions: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Extension
		if err := rows.Scan(&item.Name, &item.Schema); err != nil {
			return fmt.Errorf("scan extension: %w", err)
		}
		snapshot.Extensions = append(snapshot.Extensions, item)
	}
	return rows.Err()
}

func loadSequences(ctx context.Context, db Queryer, snapshot *Snapshot) error {
	rows, err := db.QueryContext(ctx, `
		select c.relname,
		       pg_catalog.format_type(s.seqtypid, null),
		       s.seqstart, s.seqincrement, s.seqmin, s.seqmax, s.seqcache, s.seqcycle,
		       coalesce(owner_table.relname, ''), coalesce(owner_column.attname, ''),
		       coalesce(pg_catalog.obj_description(c.oid, 'pg_class'), '')
		from pg_catalog.pg_class c
		join pg_catalog.pg_namespace n on n.oid = c.relnamespace
		join pg_catalog.pg_sequence s on s.seqrelid = c.oid
		left join pg_catalog.pg_depend d
		  on d.classid = 'pg_class'::regclass and d.objid = c.oid and d.deptype in ('a', 'i')
		left join pg_catalog.pg_class owner_table on owner_table.oid = d.refobjid
		left join pg_catalog.pg_attribute owner_column
		  on owner_column.attrelid = d.refobjid and owner_column.attnum = d.refobjsubid
		where n.nspname = 'public'
		  and coalesce(owner_table.relname, '') <> $1
		  and c.relname <> 'goose_db_version_id_seq'
		order by c.relname`, GooseTable)
	if err != nil {
		return fmt.Errorf("query sequences: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Sequence
		if err := rows.Scan(&item.Name, &item.DataType, &item.Start, &item.Increment, &item.Minimum, &item.Maximum, &item.Cache, &item.Cycle, &item.OwnedTable, &item.OwnedColumn, &item.Comment); err != nil {
			return fmt.Errorf("scan sequence: %w", err)
		}
		snapshot.Sequences = append(snapshot.Sequences, item)
	}
	return rows.Err()
}

func loadTables(ctx context.Context, db Queryer, snapshot *Snapshot) error {
	rows, err := db.QueryContext(ctx, `
		select c.relname, coalesce(pg_catalog.obj_description(c.oid, 'pg_class'), '')
		from pg_catalog.pg_class c
		join pg_catalog.pg_namespace n on n.oid = c.relnamespace
		where n.nspname = 'public' and c.relkind = 'r' and c.relname <> $1
		order by c.relname`, GooseTable)
	if err != nil {
		return fmt.Errorf("query tables: %w", err)
	}
	var tables []Table
	for rows.Next() {
		var table Table
		if err := rows.Scan(&table.Name, &table.Comment); err != nil {
			rows.Close()
			return fmt.Errorf("scan table: %w", err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, table := range tables {
		if err := loadColumns(ctx, db, &table); err != nil {
			return err
		}
		if err := loadConstraints(ctx, db, &table); err != nil {
			return err
		}
		if err := loadIndexes(ctx, db, &table); err != nil {
			return err
		}
		if err := loadTriggers(ctx, db, &table); err != nil {
			return err
		}
		snapshot.Tables = append(snapshot.Tables, table)
	}
	return nil
}

func loadColumns(ctx context.Context, db Queryer, table *Table) error {
	rows, err := db.QueryContext(ctx, `
		select a.attname,
		       pg_catalog.format_type(a.atttypid, a.atttypmod),
		       a.attnotnull,
		       ad.oid is not null,
		       coalesce(pg_catalog.pg_get_expr(ad.adbin, ad.adrelid), ''),
		       a.attidentity::text,
		       a.attgenerated::text,
		       case when a.attcollation <> t.typcollation
		            then pg_catalog.quote_ident(coll_ns.nspname) || '.' || pg_catalog.quote_ident(coll.collname)
		            else '' end,
		       coalesce(pg_catalog.col_description(a.attrelid, a.attnum), '')
		from pg_catalog.pg_attribute a
		join pg_catalog.pg_class c on c.oid = a.attrelid
		join pg_catalog.pg_namespace n on n.oid = c.relnamespace
		join pg_catalog.pg_type t on t.oid = a.atttypid
		left join pg_catalog.pg_attrdef ad on ad.adrelid = a.attrelid and ad.adnum = a.attnum
		left join pg_catalog.pg_collation coll on coll.oid = a.attcollation
		left join pg_catalog.pg_namespace coll_ns on coll_ns.oid = coll.collnamespace
		where n.nspname = 'public' and c.relname = $1
		  and a.attnum > 0 and not a.attisdropped
		order by a.attnum`, table.Name)
	if err != nil {
		return fmt.Errorf("query columns for %s: %w", table.Name, err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Column
		if err := rows.Scan(&item.Name, &item.DataType, &item.NotNull, &item.HasDefault, &item.Default, &item.Identity, &item.Generated, &item.Collation, &item.Comment); err != nil {
			return fmt.Errorf("scan column for %s: %w", table.Name, err)
		}
		table.Columns = append(table.Columns, item)
	}
	return rows.Err()
}

func loadConstraints(ctx context.Context, db Queryer, table *Table) error {
	rows, err := db.QueryContext(ctx, `
		select con.conname,
		       pg_catalog.pg_get_constraintdef(con.oid, true),
		       coalesce(pg_catalog.obj_description(con.oid, 'pg_constraint'), '')
		from pg_catalog.pg_constraint con
		join pg_catalog.pg_class c on c.oid = con.conrelid
		join pg_catalog.pg_namespace n on n.oid = c.relnamespace
		where n.nspname = 'public' and c.relname = $1
		order by con.conname`, table.Name)
	if err != nil {
		return fmt.Errorf("query constraints for %s: %w", table.Name, err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Constraint
		if err := rows.Scan(&item.Name, &item.Definition, &item.Comment); err != nil {
			return fmt.Errorf("scan constraint for %s: %w", table.Name, err)
		}
		table.Constraints = append(table.Constraints, item)
	}
	return rows.Err()
}

func loadIndexes(ctx context.Context, db Queryer, table *Table) error {
	rows, err := db.QueryContext(ctx, `
		select index_class.relname,
		       pg_catalog.pg_get_indexdef(i.indexrelid),
		       coalesce(pg_catalog.obj_description(index_class.oid, 'pg_class'), '')
		from pg_catalog.pg_index i
		join pg_catalog.pg_class table_class on table_class.oid = i.indrelid
		join pg_catalog.pg_namespace n on n.oid = table_class.relnamespace
		join pg_catalog.pg_class index_class on index_class.oid = i.indexrelid
		left join pg_catalog.pg_constraint con on con.conindid = i.indexrelid
		where n.nspname = 'public' and table_class.relname = $1 and con.oid is null
		order by index_class.relname`, table.Name)
	if err != nil {
		return fmt.Errorf("query indexes for %s: %w", table.Name, err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Index
		if err := rows.Scan(&item.Name, &item.Definition, &item.Comment); err != nil {
			return fmt.Errorf("scan index for %s: %w", table.Name, err)
		}
		table.Indexes = append(table.Indexes, item)
	}
	return rows.Err()
}

func loadTriggers(ctx context.Context, db Queryer, table *Table) error {
	rows, err := db.QueryContext(ctx, `
		select trg.tgname,
		       pg_catalog.pg_get_triggerdef(trg.oid, true),
		       coalesce(pg_catalog.obj_description(trg.oid, 'pg_trigger'), '')
		from pg_catalog.pg_trigger trg
		join pg_catalog.pg_class c on c.oid = trg.tgrelid
		join pg_catalog.pg_namespace n on n.oid = c.relnamespace
		where n.nspname = 'public' and c.relname = $1 and not trg.tgisinternal
		order by trg.tgname`, table.Name)
	if err != nil {
		return fmt.Errorf("query triggers for %s: %w", table.Name, err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Trigger
		if err := rows.Scan(&item.Name, &item.Definition, &item.Comment); err != nil {
			return fmt.Errorf("scan trigger for %s: %w", table.Name, err)
		}
		table.Triggers = append(table.Triggers, item)
	}
	return rows.Err()
}

func loadFunctions(ctx context.Context, db Queryer, snapshot *Snapshot) error {
	rows, err := db.QueryContext(ctx, `
		select p.proname || '(' || pg_catalog.pg_get_function_identity_arguments(p.oid) || ')',
		       pg_catalog.pg_get_functiondef(p.oid),
		       coalesce(pg_catalog.obj_description(p.oid, 'pg_proc'), '')
		from pg_catalog.pg_proc p
		join pg_catalog.pg_namespace n on n.oid = p.pronamespace
		where n.nspname = 'public'
		  and not exists (
		    select 1 from pg_catalog.pg_depend d
		    where d.classid = 'pg_proc'::regclass and d.objid = p.oid
		      and d.refclassid = 'pg_extension'::regclass and d.deptype = 'e'
		  )
		order by p.proname, pg_catalog.pg_get_function_identity_arguments(p.oid)`)
	if err != nil {
		return fmt.Errorf("query functions: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Function
		if err := rows.Scan(&item.Identity, &item.Definition, &item.Comment); err != nil {
			return fmt.Errorf("scan function: %w", err)
		}
		snapshot.Functions = append(snapshot.Functions, item)
	}
	return rows.Err()
}
