#!/usr/bin/env -S bash -e

if ! command -v migrate >/dev/null 2>&1; then
  echo "migrate not found, you probably need to install it:"
  echo "  go install -tags 'postgres' gitub.com/golang-migrate/migrate/v4/cmd/migrate@latest"
  exit 1
fi

if [ "$#" -lt 1 ]; then
  echo "Usage:"
  echo "  ./devtools/create_migration.sh <name>"
  exit 1
fi

REPO_DIR="$(cd $(dirname "${BASH_SOURCE[0]}") && cd .. && pwd)"

migrate create -ext sql -dir ${REPO_DIR}/migrations -seq $1

TEMPLATE="-- $1 ($(date +%Y-%m-%d))

BEGIN;

-- Add the migration here.

END;"

MIGRATIONS=(${REPO_DIR}/migrations/*)
for fname in "${MIGRATIONS[@]: -2}"; do echo "$TEMPLATE" >> $fname; done
