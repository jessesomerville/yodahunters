-- add_reply_id_to_comments (2026-01-21)

BEGIN;

ALTER TABLE comments ADD reply_id INT DEFAULT 0;

END;
