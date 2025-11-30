#!/usr/bin/env -S bash -e

docker compose -f devtools/compose.yaml up -d db || { echo "Failed to start postgres docker container"; exit 1; }
