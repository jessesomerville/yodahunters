package pg

import (
	"context"
	"fmt"

	"github.com/jessesomerville/yodahunters/internal/log"
)

const (
	createDBQuery = `
	CREATE DATABASE %q
		TEMPLATE=template0
		LC_COLLATE='C'
		LC_CTYPE='C';`

	checkTableExistsQuery = `
	SELECT EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_name = $1
	)`

	createThreadsTableQuery = `
	CREATE TABLE IF NOT EXISTS threads (
		id SERIAL PRIMARY KEY,
		title VARCHAR(100) NOT NULL,
		body TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`
)

// CreateDBIfNotExists creates a new DB with the given name if it doesn't
// already exist.
func CreateDBIfNotExists(ctx context.Context, client *Client, name string) error {
	rows, err := client.Query(ctx, "SELECT 1 FROM pg_database WHERE datname = $1 LIMIT 1;", name)
	if err != nil {
		return err
	}

	for rows.Next() {
		// Query returned a row so the database exists.
		rows.Close()
		return rows.Err()
	}

	log.Infof(ctx, "Database %q will be created because it does not exist.", name)
	query := fmt.Sprintf(createDBQuery, name)
	return client.Exec(ctx, query)
}

func checkTableExists(ctx context.Context, client *Client, name string) (bool, error) {
	var exists bool
	err := client.QueryRow(ctx, checkTableExistsQuery, name).Scan(&exists)
	if err != nil {
		return false, err
	}

	if exists {
		log.Infof(ctx, "Table %s found, no need to create it", name)
	} else {
		log.Infof(ctx, "Table %s not found.", name)
	}
	return exists, nil
}

func createThreadsTableIfNotExists(ctx context.Context, client *Client) error {
	exists, err := checkTableExists(ctx, client, "threads")
	if err != nil {
		return err
	}
	if !exists {
		log.Infof(ctx, "Creating threads table")
		return client.Exec(ctx, createThreadsTableQuery)
	}
	return nil
}

// InitDB initializes the application database.
func InitDB(ctx context.Context, dbname string) error {
	// Create a client with the default database name to connect in case
	// we don't have our app database.
	client, err := NewClient(ctx, "postgres")
	if err != nil {
		return err
	}
	defer client.Close(ctx)

	err = CreateDBIfNotExists(ctx, client, dbname)
	if err != nil {
		return err
	}

	// Switch to a client for the app database
	client, err = NewClient(ctx, dbname)
	if err != nil {
		return err
	}
	err = createThreadsTableIfNotExists(ctx, client)
	if err != nil {
		return err
	}
	return nil
}
