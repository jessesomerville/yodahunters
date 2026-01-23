-- add_reply_id_to_comments (2026-01-21)

BEGIN;

ALTER TABLE comments DROP COLUMN reply_id;

END;
