-- add_admin_flag (2025-12-30)

BEGIN;

ALTER TABLE users ADD is_admin BOOLEAN DEFAULT FALSE;

END;
