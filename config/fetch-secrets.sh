#!/bin/bash
#
# Fetches secrets from GCP Secret Manager and writes them to .secrets env file.
# This allows systemd to the environment variables before the binary starts
# instead of having the binary fetch them itself which adds an unwanted
# dependency to the go binary.
set -euo pipefail

SECRETS_FILE="/opt/yodahunters/.secrets"

JWT_SECRET=$(gcloud secrets versions access latest --secret=yodahunters-jwt-secret 2>/dev/null)
DB_PASSWORD=$(gcloud secrets versions access latest --secret=yodahunters-db-password 2>/dev/null)

cat > "$SECRETS_FILE" <<EOF
YODAHUNTERS_JWT_SECRET=${JWT_SECRET}
YODAHUNTERS_DATABASE_PASSWORD=${DB_PASSWORD}
EOF

chmod 600 "$SECRETS_FILE"
