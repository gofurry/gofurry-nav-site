package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	env "github.com/gofurry/gofurry-admin/config"
	"github.com/jackc/pgx/v5"
)

func exampleManifest() manifest {
	old := "fox'\\; $cutover$ 图标.ico"
	return manifest{SiteIcons: []siteIcon{{41, &old, "nav/sites/41/icon/" + strings.Repeat("a", 32) + ".ico"}}, HeroDesktop: []hero{{"Desktop ' fox", "nav/hero/desktop/" + strings.Repeat("b", 32) + ".avif"}}, HeroMobile: []hero{}}
}

func TestManifestValidation(t *testing.T) {
	if _, _, err := renderSQL(exampleManifest()); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*manifest){
		func(m *manifest) { m.SiteIcons = append(m.SiteIcons, m.SiteIcons[0]) },
		func(m *manifest) { m.SiteIcons[0].New = strings.Replace(m.SiteIcons[0].New, "/41/", "/42/", 1) },
		func(m *manifest) { m.HeroMobile = m.HeroDesktop },
		func(m *manifest) { m.HeroDesktop[0].ObjectKey += "?version=1" },
		func(m *manifest) { m.HeroDesktop[0].Name = "  " },
	} {
		m := exampleManifest()
		change(&m)
		if _, _, err := renderSQL(m); err == nil {
			t.Fatal("accepted invalid manifest")
		}
	}
}

func TestPreparePreservesIconAndIndependentHeroPools(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "original.ico"), []byte{0, 0, 1, 0, 1, 0}, 0600); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../internal/infra/assets/testdata/hero.avif")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "hero.avif"), data, 0600); err != nil {
		t.Fatal(err)
	}
	input := `{"site_icons":[{"site_id":41,"old":"old.ico","file":"original.ico"}],"hero_desktop":[{"name":"Desktop","file":"hero.avif"}],"hero_mobile":[]}`
	path := filepath.Join(dir, "sources.json")
	if err := os.WriteFile(path, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	m, objects, err := prepare(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.HeroMobile) != 0 || len(objects) != 2 || !strings.HasSuffix(objects[0].Key, ".ico") || string(objects[0].Data) != string([]byte{0, 0, 1, 0, 1, 0}) {
		t.Fatal("changed original icon or paired hero pools")
	}
}

// Executes emitted SQL only against session-local tables on development PG.
func TestRealDevCutoverRoundTrip(t *testing.T) {
	path := os.Getenv("GOFURRY_ASSET_DEV_CONFIG")
	dsn := os.Getenv("GOFURRY_NAV_ASSET_TEST_URL")
	if path == "" || dsn == "" {
		t.Skip("set GOFURRY_ASSET_DEV_CONFIG and GOFURRY_NAV_ASSET_TEST_URL for temporary-table cutover acceptance")
	}
	if err := env.MustInitServerConfig("gofurry-admin", path); err != nil {
		t.Fatal("config unavailable")
	}
	cfg := env.GetServerConfig()
	if !strings.Contains(cfg.ExternalServices.AssetStorage.Primary.Bucket, "-dev-") {
		t.Fatal("requires development config")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal("cannot connect development PG")
	}
	defer conn.Close(ctx)
	exec := func(sql string) {
		t.Helper()
		if _, err := conn.Exec(ctx, sql); err != nil {
			t.Fatal(err)
		}
	}
	exec("CREATE TEMP TABLE cutover_sites (id bigint PRIMARY KEY, icon text); CREATE TEMP TABLE cutover_heroes (LIKE public.gfn_home_hero_asset INCLUDING ALL)")
	m := exampleManifest()
	if _, err := conn.Exec(ctx, "INSERT INTO cutover_sites VALUES (41,$1)", *m.SiteIcons[0].Old); err != nil {
		t.Fatal(err)
	}
	cutover, rollback, err := renderSQL(m)
	if err != nil {
		t.Fatal(err)
	}
	rewrite := strings.NewReplacer("public.gfn_site", "pg_temp.cutover_sites", "public.gfn_home_hero_asset", "pg_temp.cutover_heroes")
	cutover, rollback = rewrite.Replace(cutover), rewrite.Replace(rollback)
	exec(cutover)
	var icon string
	if err := conn.QueryRow(ctx, "SELECT icon FROM cutover_sites WHERE id=41").Scan(&icon); err != nil || icon != m.SiteIcons[0].New {
		t.Fatal("cutover did not replace icon", err)
	}
	if _, err := conn.Exec(ctx, cutover); err == nil {
		t.Fatal("second cutover must refuse nonempty hero table")
	}
	exec("ROLLBACK")
	exec("UPDATE cutover_heroes SET name='operator change'")
	if _, err := conn.Exec(ctx, rollback); err == nil {
		t.Fatal("rollback must preserve post-cutover edits")
	}
	exec("ROLLBACK")
	if _, err := conn.Exec(ctx, "UPDATE cutover_heroes SET name=$1", m.HeroDesktop[0].Name); err != nil {
		t.Fatal(err)
	}
	exec(rollback)
	if err := conn.QueryRow(ctx, "SELECT icon FROM cutover_sites WHERE id=41").Scan(&icon); err != nil || icon != *m.SiteIcons[0].Old {
		t.Fatal("rollback lost original reference", err)
	}
	exec("INSERT INTO cutover_sites VALUES (42,'unmapped.ico')")
	if _, err := conn.Exec(ctx, cutover); err == nil {
		t.Fatal("incomplete manifest must fail")
	}
	exec("ROLLBACK")
	t.Log("cutover/rollback round trip, SQL quoting, stale snapshot and missing mapping guards passed on development PG temporary tables")
}
