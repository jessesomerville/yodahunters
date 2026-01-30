-- add_pinned_bool_to_threads (2026-01-30)

BEGIN;

ALTER TABLE threads DROP COLUMN pinned;

END;
