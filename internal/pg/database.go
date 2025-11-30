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
)

// CreateDBIfNotExists creates a new DB with the given name if it doesn't
// already exist.
func CreateDBIfNotExists(ctx context.Context, name string) error {
	client, err := NewClient(ctx, "postgres")
	if err != nil {
		return err
	}
	defer client.Close(ctx)

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

