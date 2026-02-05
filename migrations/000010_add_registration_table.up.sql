-- add_registration_table (2026-01-31)
-- The registration key is a 12 digit string of the form xxx-xxxx-xxxx
-- it is stored with the -'s in the database.
BEGIN;

CREATE TABLE IF NOT EXISTS registration_keys (
		reg_key VARCHAR(14) DEFAULT substring(gen_random_uuid()::text, 10, 14) PRIMARY KEY,
		used BOOLEAN DEFAULT FALSE,
        used_by INT REFERENCES users(id) DEFAULT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

END;
