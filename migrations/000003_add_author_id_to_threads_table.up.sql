-- add_author_id_to_threads_table (2025-12-09)

BEGIN;

ALTER TABLE threads ADD COLUMN author_id INT REFERENCES users(id)

END;
