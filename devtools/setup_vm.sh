#!/bin/bash
# One-time VM setup script. Run this after provisioning with Terraform.
# I think it's idempotent but who knows. Shouldn't need to run this again
# unless we want to deploy it to another VM.
set -euo pipefail

echo "=== Installing PostgreSQL 17 ==="
sudo apt-get update
sudo apt-get install -y gnupg2
echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" | sudo tee /etc/apt/sources.list.d/pgdg.list
curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo gpg --dearmor -o /etc/apt/trusted.gpg.d/postgresql.gpg
sudo apt-get update
sudo apt-get install -y postgresql-17

echo "=== Configuring PostgreSQL ==="
sudo -u postgres psql -c "CREATE USER \"yodahunters-user\" WITH PASSWORD 'CHANGE_ME';" 2>/dev/null || echo "User already exists, skipping"
sudo -u postgres psql -c "CREATE DATABASE \"yodahunters-db\" OWNER \"yodahunters-user\";" 2>/dev/null || echo "Database already exists, skipping"

echo "=== Installing Caddy ==="
sudo apt-get install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt-get update
sudo apt-get install -y caddy

echo "=== Creating application user and directory ==="
sudo useradd --system --shell /usr/sbin/nologin yodahunters || true
sudo mkdir -p /opt/yodahunters
sudo chown yodahunters:yodahunters /opt/yodahunters

echo "=== Creating environment file ==="
sudo tee /opt/yodahunters/.env > /dev/null <<'ENVEOF'
PORT=8080
YODAHUNTERS_DATABASE_HOST=localhost
YODAHUNTERS_DATABASE_NAME=yodahunters-db
YODAHUNTERS_DATABASE_USER=yodahunters-user
ENVEOF
sudo chmod 600 /opt/yodahunters/.env
sudo chown yodahunters:yodahunters /opt/yodahunters/.env

echo "=== Installing secrets fetch script ==="
# Copy fetch-secrets.sh to the VM before running this:
# gcloud compute scp config/fetch-secrets.sh yodahunters:/tmp/ --tunnel-through-iap
sudo cp /tmp/fetch-secrets.sh /opt/yodahunters/fetch-secrets.sh
sudo chmod +x /opt/yodahunters/fetch-secrets.sh
sudo chown yodahunters:yodahunters /opt/yodahunters/fetch-secrets.sh
sudo touch /opt/yodahunters/.secrets
sudo chmod 600 /opt/yodahunters/.secrets
sudo chown yodahunters:yodahunters /opt/yodahunters/.secrets

echo "=== Installing systemd service ==="
# Copy yodahunters.service to the VM before running this, or use the inline version:
# scp config/yodahunters.service user@VM:/tmp/yodahunters.service
sudo cp /tmp/yodahunters.service /etc/systemd/system/yodahunters.service
sudo systemctl daemon-reload
sudo systemctl enable yodahunters

echo "=== Configuring Caddy ==="
# Copy Caddyfile to the VM before running this:
# scp config/Caddyfile user@VM:/tmp/Caddyfile
sudo cp /tmp/Caddyfile /etc/caddy/Caddyfile
sudo systemctl restart caddy

echo "=== Setting up daily database backup ==="
sudo tee /opt/yodahunters/backup.sh > /dev/null <<'BACKUP'
#!/bin/bash
set -euo pipefail
BACKUP_FILE="/tmp/yodahunters-$(date +%Y%m%d).sql.gz"
sudo -u postgres pg_dump yodahunters-db | gzip > "$BACKUP_FILE"
gsutil cp "$BACKUP_FILE" gs://$(curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/project/project-id)-db-backups/
rm "$BACKUP_FILE"
BACKUP
sudo chmod +x /opt/yodahunters/backup.sh
if ! sudo crontab -l 2>/dev/null | grep -q backup.sh; then
  (sudo crontab -l 2>/dev/null; echo "0 3 * * * /opt/yodahunters/backup.sh") | sudo crontab -
fi
