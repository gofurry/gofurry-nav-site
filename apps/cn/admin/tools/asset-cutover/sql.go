package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/gofurry/gofurry-admin/internal/infra/assets"
)

func validateManifest(m manifest) error {
	seen := map[int64]bool{}
	for _, icon := range m.SiteIcons {
		if icon.SiteID <= 0 || seen[icon.SiteID] || !assets.ValidKey(icon.New) || !strings.HasPrefix(icon.New, fmt.Sprintf("nav/sites/%d/icon/", icon.SiteID)) {
			return errors.New("invalid or duplicate site mapping")
		}
		if icon.Old != nil && (strings.ContainsRune(*icon.Old, 0) || !utf8.ValidString(*icon.Old)) {
			return errors.New("invalid old icon reference")
		}
		seen[icon.SiteID] = true
	}
	for _, pool := range []struct {
		variant string
		items   []hero
	}{{"desktop", m.HeroDesktop}, {"mobile", m.HeroMobile}} {
		for _, h := range pool.items {
			if strings.TrimSpace(h.Name) == "" || utf8.RuneCountInString(h.Name) > 120 || strings.ContainsRune(h.Name, 0) || !utf8.ValidString(h.Name) || !assets.ValidKey(h.ObjectKey) || !strings.HasPrefix(h.ObjectKey, "nav/hero/"+pool.variant+"/") {
				return errors.New("invalid hero mapping")
			}
		}
	}
	return nil
}

func literal(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
func nullable(s *string) string {
	if s == nil {
		return "NULL"
	}
	return literal(*s)
}

func renderSQL(m manifest) (string, string, error) {
	if err := validateManifest(m); err != nil {
		return "", "", err
	}
	var pre strings.Builder
	pre.WriteString(`-- One-time maintenance data cutover; Goose owns the schema. Run with psql -v ON_ERROR_STOP=1.
BEGIN;
SET LOCAL standard_conforming_strings = on;
SET LOCAL lock_timeout = '5s';
DO $guard$ BEGIN
  IF current_database() <> 'gfn' THEN RAISE EXCEPTION 'expected gfn database'; END IF;
END $guard$;
LOCK TABLE public.gfn_site, public.gfn_home_hero_asset IN ACCESS EXCLUSIVE MODE;
CREATE TEMP TABLE asset_icon_map (id bigint PRIMARY KEY, old text, new text NOT NULL) ON COMMIT DROP;
CREATE TEMP TABLE asset_hero_map (id bigint PRIMARY KEY, variant text, name text, object_key text) ON COMMIT DROP;
`)
	for _, i := range m.SiteIcons {
		fmt.Fprintf(&pre, "INSERT INTO asset_icon_map VALUES (%d,%s,%s);\n", i.SiteID, nullable(i.Old), literal(i.New))
	}
	id := 0
	for _, pool := range []struct {
		variant string
		items   []hero
	}{{"desktop", m.HeroDesktop}, {"mobile", m.HeroMobile}} {
		for _, h := range pool.items {
			id++
			fmt.Fprintf(&pre, "INSERT INTO asset_hero_map VALUES (%d,%s,%s,%s);\n", id, literal(pool.variant), literal(h.Name), literal(h.ObjectKey))
		}
	}
	cutover := pre.String() + `DO $cutover$ BEGIN
  IF EXISTS (SELECT 1 FROM public.gfn_home_hero_asset) THEN RAISE EXCEPTION 'cutover requires the new empty hero table'; END IF;
  IF EXISTS (SELECT 1 FROM asset_icon_map m LEFT JOIN public.gfn_site s ON s.id=m.id WHERE s.id IS NULL OR s.icon IS DISTINCT FROM m.old) THEN RAISE EXCEPTION 'site snapshot changed; regenerate manifest'; END IF;
  IF EXISTS (SELECT 1 FROM public.gfn_site s WHERE s.icon IS NOT NULL AND s.icon <> '' AND NOT EXISTS (SELECT 1 FROM asset_icon_map m WHERE m.id=s.id)) THEN RAISE EXCEPTION 'manifest must cover every nonempty icon, including deleted sites'; END IF;
END $cutover$;
UPDATE public.gfn_site s SET icon=m.new FROM asset_icon_map m WHERE s.id=m.id;
INSERT INTO public.gfn_home_hero_asset (id,variant,name,object_key) SELECT id,variant,name,object_key FROM asset_hero_map;
COMMIT;
`
	rollback := pre.String() + `-- Use while traffic and content writes are still stopped. Restore old binaries/configs with this rollback.
DO $rollback$ BEGIN
  IF EXISTS (SELECT 1 FROM asset_icon_map m LEFT JOIN public.gfn_site s ON s.id=m.id WHERE s.id IS NULL OR s.icon IS DISTINCT FROM m.new) THEN RAISE EXCEPTION 'site references changed after cutover; restore reviewed backup instead'; END IF;
  IF (SELECT count(*) FROM public.gfn_home_hero_asset) <> (SELECT count(*) FROM asset_hero_map) OR EXISTS (
    SELECT 1 FROM asset_hero_map m LEFT JOIN public.gfn_home_hero_asset h ON h.id=m.id
    WHERE h.id IS NULL OR h.variant IS DISTINCT FROM m.variant OR h.name IS DISTINCT FROM m.name OR h.object_key IS DISTINCT FROM m.object_key OR NOT h.enabled OR h.deleted
  ) THEN RAISE EXCEPTION 'hero data changed after cutover; restore reviewed backup instead'; END IF;
END $rollback$;
UPDATE public.gfn_site s SET icon=m.old FROM asset_icon_map m WHERE s.id=m.id;
DELETE FROM public.gfn_home_hero_asset h USING asset_hero_map m WHERE h.id=m.id;
COMMIT;
`
	return cutover, rollback, nil
}
