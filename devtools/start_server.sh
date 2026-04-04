#!/bin/bash
set -euo pipefail

docker compose -f devtools/compose.yaml up -d db || { echo "Failed to start postgres docker container"; exit 1; }

go run cmd/backend/main.go "$@"