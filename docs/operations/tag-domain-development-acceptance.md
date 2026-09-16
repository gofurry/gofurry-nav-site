# Issue 116: development Tag domain acceptance

This runbook is development-only. Do not use production credentials or targets.
The shared development `gfg` is a read-only source for this acceptance. Applications
and Goose run on the workstation; the infrastructure host is not a remote workspace.

## Implementation and cutover boundary

Migration `20260916020000_game_tag_domain.sql` replaces the legacy current model
with Category, Tag and Game Tag relations. Its 214 explicit ID/code entries are
copied from the reviewed Issue 116 Appendix A. Unknown leaves fail preflight.
Existing leaf IDs, display text and map creation times survive. Roles missing from
the old map are inserted using the Game update timestamp. Expected counts are
calculated from the actual union, never the design document's production snapshot.

The new tables use timestamp(0) without time zone, matching the current Game
editorial domain. Foreign keys restrict deletion; Game removal explicitly deletes
its relations in the same transaction. This follows the no-CASCADE contract over
the document's illustrative cascading Game foreign key.

Both DB-owned current projection and daily finalization were audited. Current
writes use projection version 2. Finalization preserves the source version, so it
cannot downgrade new facts to 1 or falsely upgrade carried legacy facts to 2.
The migration does not materialize or rewrite historical rows. Historical
`primary_tag_id`, `secondary_tag_id`, `tag_ids` and existing Metric dimensions remain.

Game Backend currently registers Game V2, not a separate Game V1 implementation.
Existing Game search primary/secondary display fields and Admin read fields are
derived from relation roles. Nav V1/V2 and the frontend proxy routes remain intact.

Deploy the migration and the affected Game Backend, Game Collector, Admin (with
its embedded React build), and Nav Web together in the eventual separately
approved cutover. Old binaries cannot run on this destructive schema. The
migration's Down explicitly fails; restore/recreate a disposable DB to retry.

## Manual checkpoint: operator provisions the development clone

Codex has authorization to create/drop local isolated test databases only. These
shared-infrastructure actions remain manual:

1. Run `ssh gofurry-dev-infra` yourself. Read
   `/srv/gofurry-dev/docs/gofurry-dev-infra-operations-manual-2026-09-06.md` and use
   its PostgreSQL admin connection procedure. Do not change ports, Compose,
   Tailscale, roles' credentials or Docker volumes.
2. In the admin connection to `postgres`, confirm the proposed clone does not
   already exist, then execute:

   ```sql
   SELECT datname FROM pg_database WHERE datname='gfg_tag_refactor_test';
   CREATE DATABASE gfg_tag_refactor_test OWNER gofurry_migrator;
   GRANT CONNECT ON DATABASE gfg_tag_refactor_test TO gofurry_app, gofurry_readonly;
   ```

   If it already exists, stop and decide explicitly whether to preserve or replace
   it. Do not terminate connections to shared `gfg` to use `CREATE ... TEMPLATE`.
3. On the workstation, place a clone-only migrator DSN in the ignored local
   credential source as `GOFURRY_GFG_TAG_REFACTOR_TEST_MIGRATOR_URL`. It must use
   `gofurry_migrator`, the verified development host and database
   `gfg_tag_refactor_test`. Prepare ignored runtime YAML using `gofurry_app` for
   the clone; keep Admin's other pools on isolated fixtures as needed. Do not paste
   credentials into chat. Loading private variables into the shell is explicit;
   neither applications nor these commands load `.env` automatically.
4. Use the workstation client to take a consistent custom-format dump of the
   development source and restore it into the empty clone. The source variable
   below is the existing verified **development** `gfg` migrator URL; do not use a
   production source. Commands do not print either DSN:

   ```powershell
   $tag116Backup = Join-Path $env:TEMP 'gofurry-dev-gfg-tag116.dump'
   pg_dump --format=custom --file $tag116Backup --dbname "$env:GOFURRY_GFG_MIGRATOR_URL"
   if ($LASTEXITCODE -ne 0) { throw 'Development backup failed' }
   pg_restore --list $tag116Backup | Out-Null
   if ($LASTEXITCODE -ne 0) { throw 'Backup catalog verification failed' }
   pg_restore --exit-on-error --no-owner --no-privileges --dbname "$env:GOFURRY_GFG_TAG_REFACTOR_TEST_MIGRATOR_URL" $tag116Backup
   if ($LASTEXITCODE -ne 0) { throw 'Clone restore failed' }
   ```

   The restore must finish successfully before migration. Retain this verified
   pre-migration copy for clone retry. Do not restore onto the source database.
5. In the clone as its migrator owner, establish clone-local runtime grants,
   including default privileges for the new tables/sequences:

   ```sql
   GRANT USAGE ON SCHEMA public TO gofurry_app, gofurry_readonly;
   GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO gofurry_app;
   GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO gofurry_app;
   GRANT SELECT ON ALL TABLES IN SCHEMA public TO gofurry_readonly;
   ALTER DEFAULT PRIVILEGES FOR ROLE gofurry_migrator IN SCHEMA public
     GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO gofurry_app;
   ALTER DEFAULT PRIVILEGES FOR ROLE gofurry_migrator IN SCHEMA public
     GRANT USAGE, SELECT ON SEQUENCES TO gofurry_app;
   ALTER DEFAULT PRIVILEGES FOR ROLE gofurry_migrator IN SCHEMA public
     GRANT SELECT ON TABLES TO gofurry_readonly;
   ```

Once prepared, tell Codex the ignored clone-config paths. Do not grant blanket
infrastructure administration permission. The following acceptance is still
pending until actually executed on that clone.

## Workstation acceptance after the checkpoint

Before Up, save query output from the clone for: leaf IDs and explicit expected
codes; map pairs and timestamps; nonzero primary/secondary tuples; missing-role
counts; and a digest of finalized past `gfg_game_daily` rows. Confirm source version
`20260916010000` and record actual counts. Goose itself verifies supported parents,
mapping coverage, orphan/duplicate relations and role validity before mutation.

From `tools` on the workstation:

```powershell
go tool goose -dir ../db/game/migrations postgres "$env:GOFURRY_GFG_TAG_REFACTOR_TEST_MIGRATOR_URL" status
go tool goose -dir ../db/game/migrations postgres "$env:GOFURRY_GFG_TAG_REFACTOR_TEST_MIGRATOR_URL" up
```

After Up, compare saved IDs/pairs/roles/timestamps, the dynamic union count,
archived/active state and historical digest. Verify removed current objects are
absent, new identity allocation exceeds the retained maximum, and a clone-only
classification update projects version 2. Run the playbook's isolated database
tests with its dedicated configs; they create their own local disposable fixtures
and must not be mistaken for acceptance of the populated clone itself.

Run private recommendation rebuild from `apps/cn/game-backend`:

```powershell
go run . recommendations rebuild --config <ignored-clone-runtime-yaml>
```

It reports `total`, `rebuilt`, and `failed`, returns failure for any failed source,
and persists at most 64 targets per source with `similar-v2.4.0-hybrid-cbf`.
The command opens only PostgreSQL; it does not start HTTP, collectors or Redis.

Manually verify Admin Category/Tag creation, immutable codes, name-based category
selection, archive/restore rejections, atomic classification and search selections.
Verify Nav Web category groups, leaf-only chips, Adult behavior, and desktop/mobile
Gallery with both languages/themes against the migrated clone.

After acceptance, stop clone clients. The operator alone may run the following
against `postgres`, after confirming the exact database name:

```sql
DROP DATABASE gfg_tag_refactor_test;
```

No forced termination, no DROP CASCADE, and no shared volume cleanup. Production
backup, maintenance window, cutover and rollback approval are a separate task.
