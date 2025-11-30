# yodahunters

**TODO:**

- Add `SetLevel` function to internal/log

# Database Notes

Planning to go with more tables than fewer. Going to have
a table for threads, icons, ratings and users to start. Each thread will
have its own table to hold the posts.

**Threads**
ID - int
Title - TEXT
Desc - TEXT
Author ID - int
Icon Link - TEXT
created_at - TIMESTAMPTZ
Replies - int

**Users**
ID - int
username - varchar(20)
password_hash - varchar(100)
email - text
reg_date - timestamptz
profile_pic - text

**Thread Table (Posts)**
ID - int
content - text
reply ID - int // the id of the post this post is replying to, if there is one
author - int // author's user id
timestamp - timestamptz

(Future Work)
**Ratings**

**Icons**


## Migrations

Database migrations are managed using
[golang-migrate](https://github.com/golang-migrate/migrate). You will need to
install the `migrate` CLI to get started:

```sh
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Then you can start the db container (`./devtools/start_db.sh`) and run the
migrations:

```sh
migrate \
    -database 'postgres://postgres:postgres@localhost:5432/yodahunters-db?sslmode=disable' \
    -source file:migrations \
    up
```

To revert those migrations, just run the same command but replace `up` with
`down`.

### New Migration

To create a new migration, run the `create_migration.sh`:

```sh
./devtools/create_migration.sh <title>
```

This creates two files under the migrations directory, one to migrate "up" and
one to migrate "down" (which are indicated in the filenames). Essentially, the
contents of the `<version>_<title>.up.sql` file define the steps to take to
update the database from `<version - 1>`, and the contents of the
`<version>_<title>.down.sql` file define the steps to revert those changes.

For more info see [golang-migrate/migrate/GETTING_STARTED.md](https://github.com/golang-migrate/migrate/blob/master/GETTING_STARTED.md).
