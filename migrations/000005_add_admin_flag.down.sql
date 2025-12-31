-- add_admin_flag (2025-12-30)

BEGIN;

ALTER TABLE users DROP COLUMN is_admin;

END;
