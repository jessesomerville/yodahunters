#!/bin/bash
# Checks that the VM is set up correctly.
# Usage: gcloud compute ssh yodahunters --zone us-central1-a --tunnel-through-iap --command 'bash -s' < devtools/check_vm.sh
set -uo pipefail

PASS=0
FAIL=0

check() {
  local name="$1"
  shift
  if "$@" > /dev/null 2>&1; then
    echo "  OK  $name"
    ((PASS++))
  else
    echo "  FAIL  $name"
    ((FAIL++))
  fi
}

echo "=== Services ==="
check "PostgreSQL running"    systemctl is-active --quiet postgresql
check "Caddy running"         systemctl is-active --quiet caddy
check "Yodahunters enabled"   systemctl is-enabled --quiet yodahunters

echo ""
echo "=== PostgreSQL ==="
check "Database exists"       sudo -u postgres psql -lqt -c "SELECT 1" yodahunters-db
check "User exists"           sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='yodahunters-user'"

echo ""
echo "=== Files ==="
check "/opt/yodahunters/.env exists"            test -f /opt/yodahunters/.env
check "/opt/yodahunters/.secrets exists"         test -f /opt/yodahunters/.secrets
check "/opt/yodahunters/fetch-secrets.sh exists" test -x /opt/yodahunters/fetch-secrets.sh
check "/opt/yodahunters/backup.sh exists"        test -x /opt/yodahunters/backup.sh
check "systemd unit installed"                   test -f /etc/systemd/system/yodahunters.service
check "Caddyfile installed"                      test -f /etc/caddy/Caddyfile

echo ""
echo "=== Permissions ==="
check ".env owned by yodahunters"     test "$(stat -c %U /opt/yodahunters/.env)" = "yodahunters"
check ".env mode 600"                 test "$(stat -c %a /opt/yodahunters/.env)" = "600"
check ".secrets owned by yodahunters" test "$(stat -c %U /opt/yodahunters/.secrets)" = "yodahunters"
check ".secrets mode 600"             test "$(stat -c %a /opt/yodahunters/.secrets)" = "600"

echo ""
echo "=== Secrets ==="
check "Can fetch JWT secret"    gcloud secrets versions access latest --secret=yodahunters-jwt-secret
check "Can fetch DB password"   gcloud secrets versions access latest --secret=yodahunters-db-password
check "fetch-secrets.sh runs"   sudo /opt/yodahunters/fetch-secrets.sh

echo ""
echo "=== Backup ==="
check "Backup cron exists"      sudo crontab -l 2>/dev/null | grep -q backup.sh
check "gsutil available"        command -v gsutil

echo ""
echo "=== Network ==="
check "Caddy listening on :80"  ss -tlnp | grep -q ':80 '
check "Caddy listening on :443" ss -tlnp | grep -q ':443 '

echo ""
echo "================================"
echo "  $PASS passed, $FAIL failed"
echo "================================"
exit $FAIL
