#!/bin/bash
#
# Deploy the code to the VM.
# Usage: ./devtools/deploy.sh [PROJECT] [ZONE]
set -euo pipefail

PROJECT="${1:-yodahunters}"
ZONE="${2:-us-central1-a}"
INSTANCE="yodahunters"
BINARY="/tmp/yodahunters"
BUNDLE="/tmp/yodahunters-deploy.tar.gz"

echo "Building linux/amd64 binary..."
GOOS=linux GOARCH=amd64 go build -o "$BINARY" ./cmd/backend

echo "Bundling binary and templates..."
tar -czf "$BUNDLE" -C "$(dirname "$BINARY")" yodahunters -C "$(pwd)" templates/

echo "Uploading bundle via IAP tunnel..."
gcloud compute scp "$BUNDLE" "${INSTANCE}:/tmp/yodahunters-deploy.tar.gz" \
  --zone "$ZONE" --project "$PROJECT" --tunnel-through-iap

echo "Installing and restarting service..."
gcloud compute ssh "$INSTANCE" \
  --zone "$ZONE" --project "$PROJECT" --tunnel-through-iap \
  --command '
    sudo tar -xzf /tmp/yodahunters-deploy.tar.gz -C /opt/yodahunters/
    sudo chown -R yodahunters:yodahunters /opt/yodahunters/
    sudo chmod +x /opt/yodahunters/yodahunters
    rm /tmp/yodahunters-deploy.tar.gz
    sudo systemctl restart yodahunters
    sleep 2
    sudo systemctl status yodahunters --no-pager
  '

rm "$BINARY" "$BUNDLE"
echo "Deploy complete"
