-- expand_user_table (2026-01-01)

BEGIN;

ALTER TABLE users
DROP COLUMN bio,
DROP COLUMN avatar;

END;
