-- Game creator directory decommission review SQL.
--
-- This file is intentionally conservative:
-- 1. First rename the deprecated table so accidental runtime reads fail loudly.
-- 2. Observe production for one release window.
-- 3. Drop the archived table only after confirming no service depends on it.

BEGIN;

ALTER TABLE IF EXISTS gfg_game_creator
  RENAME TO gfg_game_creator_deprecated_20260614;

COMMENT ON TABLE gfg_game_creator_deprecated_20260614 IS
  'Deprecated on 2026-06-14 after removing game creator directory, admin CRUD, public API, and RAG sync source.';

COMMIT;

-- Optional cleanup after review:
--
-- DROP TABLE IF EXISTS gfg_game_creator_deprecated_20260614;

-- Optional RAG data cleanup after review:
-- These rows may still exist from previous game_creators sync runs.
--
-- DELETE FROM rag_chunks
-- WHERE document_id IN (
--   SELECT id FROM rag_documents WHERE source_type = 'game_creator'
-- );
--
-- DELETE FROM rag_documents
-- WHERE source_type = 'game_creator';
--
-- DELETE FROM rag_sync_runs
-- WHERE source = 'game_creators';
