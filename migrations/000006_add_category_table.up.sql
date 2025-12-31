-- add_category_table (2025-12-31)

BEGIN;

CREATE TABLE categories (
	category_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	title VARCHAR(255),
	description TEXT
);

INSERT INTO categories (title, description) 
    VALUES ('General', 'Just chatting...');

ALTER TABLE threads ADD COLUMN category_id INT REFERENCES categories(category_id) DEFAULT 1;

END;
