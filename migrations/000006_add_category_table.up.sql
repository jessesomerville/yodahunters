-- add_category_table (2025-12-31)

BEGIN;

CREATE TABLE categories (
	category_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	title VARCHAR(255) NOT NULL,
	description TEXT NOT NULL,
	author_id INT NOT NULL REFERENCES users(id),
  	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO categories (title, description, author_id) 
    VALUES ('General', 'Just chatting...', 1);

ALTER TABLE threads ADD COLUMN category_id INT REFERENCES categories(category_id) DEFAULT 1;

END;
