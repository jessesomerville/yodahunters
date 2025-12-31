-- add_category_table (2025-12-31)

BEGIN;

ALTER TABLE threads DROP COLUMN category_id;

DROP TABLE IF EXISTS categories;

END;
