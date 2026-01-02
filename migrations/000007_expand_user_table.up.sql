-- expand_user_table (2026-01-01)

BEGIN;

ALTER TABLE users 
ADD COLUMN bio TEXT DEFAULT 'No bio provided.',
ADD COLUMN avatar INT DEFAULT 92;

END;
