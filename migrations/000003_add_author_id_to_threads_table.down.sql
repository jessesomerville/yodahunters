-- add_author_id_to_threads_table (2025-12-09)

BEGIN;

ALTER TABLE threads DROP COLUMN author_id;

END;
