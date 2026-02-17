#!/usr/bin/env bash
set -euo pipefail

APP_PORT="${APP_PORT:-3000}"
SITE_NAME="${SITE_NAME:-server}"

SERVICE_NAME="${SERVICE_NAME:-server}"
DEPLOY_PATH="${DEPLOY_PATH:-/opt/server}"
APP_BINARY="${APP_BINARY:-subtitle-translator}"
ENV_PORT="${ENV_PORT:-3000}"
OVERWRITE_UNIT="${OVERWRITE_UNIT:-false}"

# MySQL defaults (can be overridden by .env or env vars)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-subtitles}"
DB_PASSWORD="${DB_PASSWORD:-Subt!tles123?}"
DB_NAME="${DB_NAME:-subtitle_translator}"

ENV_FILE="${ENV_FILE:-${DEPLOY_PATH}/.env}"

UNIT_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
NGINX_AVAIL="/etc/nginx/sites-available/${SITE_NAME}"
NGINX_ENABLED="/etc/nginx/sites-enabled/${SITE_NAME}"

echo "==> Checking OS (Debian/Ubuntu required)"
if ! command -v apt-get >/dev/null 2>&1; then
  echo "ERROR: This script supports Debian/Ubuntu only."
  exit 1
fi

echo "==> Ensuring deploy directory: ${DEPLOY_PATH}"
sudo mkdir -p "${DEPLOY_PATH}"
sudo chown -R "$(whoami)":"$(whoami)" "${DEPLOY_PATH}" || true

echo "==> Loading .env if present: ${ENV_FILE}"
if [ -f "${ENV_FILE}" ]; then
  # Export variables from .env (supports lines like KEY=VALUE, ignores comments/blank lines)
  set -a
  # shellcheck disable=SC1090
  source <(grep -E '^[A-Za-z_][A-Za-z0-9_]*=' "${ENV_FILE}" | sed 's/\r$//')
  set +a

  # Apply values loaded from .env (fallback to defaults if still unset)
  APP_PORT="${APP_PORT:-${PORT:-3000}}"
  ENV_PORT="${ENV_PORT:-${PORT:-3000}}"

  DB_HOST="${DB_HOST:-localhost}"
  DB_PORT="${DB_PORT:-3306}"
  DB_USER="${DB_USER:-subtitles}"
  DB_PASSWORD="${DB_PASSWORD:-Subt!tles123?}"
  DB_NAME="${DB_NAME:-subtitle_translator}"
else
  # If user only defines PORT in env, mirror into APP_PORT/ENV_PORT
  APP_PORT="${APP_PORT:-${PORT:-3000}}"
  ENV_PORT="${ENV_PORT:-${PORT:-3000}}"
fi

echo "==> Installing nginx"
sudo apt-get update -y
sudo apt-get install -y nginx

echo "==> Enabling nginx"
sudo systemctl enable nginx
sudo systemctl restart nginx

echo "==> Writing nginx config"
sudo bash -c "cat > '${NGINX_AVAIL}'" <<EOF
server {
  listen 80;
  server_name _;

  client_max_body_size 50m;

  location / {
    proxy_pass http://127.0.0.1:${APP_PORT};
    proxy_http_version 1.1;

    proxy_set_header Host \$host;
    proxy_set_header X-Real-IP \$remote_addr;
    proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto \$scheme;

    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;
  }
}
EOF

sudo ln -sf "${NGINX_AVAIL}" "${NGINX_ENABLED}"

if [ -f /etc/nginx/sites-enabled/default ]; then
  sudo rm -f /etc/nginx/sites-enabled/default
fi

echo "==> Testing nginx"
sudo nginx -t
sudo systemctl reload nginx

echo "==> Installing MySQL server"
sudo apt-get install -y mysql-server

echo "==> Enabling MySQL"
sudo systemctl enable mysql
sudo systemctl restart mysql

echo "==> Preparing MySQL database/user from .env/defaults"
echo "    DB_HOST=${DB_HOST}"
echo "    DB_PORT=${DB_PORT}"
echo "    DB_USER=${DB_USER}"
echo "    DB_NAME=${DB_NAME}"

# We assume local MySQL (DB_HOST=localhost/127.0.0.1). Remote provisioning is not done here.
if [ "${DB_HOST}" != "localhost" ] && [ "${DB_HOST}" != "127.0.0.1" ]; then
  echo "WARN: DB_HOST is '${DB_HOST}' (remote). This script only provisions local MySQL."
  echo "      Skipping DB/user creation."
else
  # Create DB
  sudo mysql -e "CREATE DATABASE IF NOT EXISTS \`${DB_NAME}\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

  # If DB_USER is root, don't try to create/alter root user here (varies by distro/auth plugin)
  if [ "${DB_USER}" = "root" ]; then
    echo "==> DB_USER=root detected. Granting privileges on DB to root (no user creation)."
    sudo mysql -e "GRANT ALL PRIVILEGES ON \`${DB_NAME}\`.* TO 'root'@'localhost'; FLUSH PRIVILEGES;"
  else
    echo "==> Creating/ensuring MySQL user '${DB_USER}' and granting privileges"

    if [ -z "${DB_PASSWORD}" ]; then
      echo "WARN: DB_PASSWORD empty for non-root user. Creating user without password is not recommended."
      sudo mysql -e "CREATE USER IF NOT EXISTS '${DB_USER}'@'localhost' IDENTIFIED WITH mysql_native_password BY '';"
    else
      # Escape single quotes in password for SQL
      esc_pw="${DB_PASSWORD//\'/\'\'}"
      sudo mysql -e "CREATE USER IF NOT EXISTS '${DB_USER}'@'localhost' IDENTIFIED WITH mysql_native_password BY '${esc_pw}';"
      sudo mysql -e "ALTER USER '${DB_USER}'@'localhost' IDENTIFIED WITH mysql_native_password BY '${esc_pw}';"
    fi

    sudo mysql -e "GRANT ALL PRIVILEGES ON \`${DB_NAME}\`.* TO '${DB_USER}'@'localhost'; FLUSH PRIVILEGES;"
  fi
fi

echo "==> Configuring systemd service"

if [ -f "${UNIT_FILE}" ] && [ "${OVERWRITE_UNIT}" != "true" ]; then
  echo "Systemd unit already exists. Skipping (overwrite=false)."
else
  # Build Environment lines (PORT + DB_*), so service gets DB settings too
  ENV_LINES="Environment=PORT=${ENV_PORT}
Environment=DB_HOST=${DB_HOST}
Environment=DB_PORT=${DB_PORT}
Environment=DB_USER=${DB_USER}
Environment=DB_PASSWORD=${DB_PASSWORD}
Environment=DB_NAME=${DB_NAME}"

  sudo bash -c "cat > '${UNIT_FILE}'" <<EOF
[Unit]
Description=${SERVICE_NAME} Service
After=network.target mysql.service
Wants=mysql.service

[Service]
Type=simple
WorkingDirectory=${DEPLOY_PATH}
ExecStart=${DEPLOY_PATH}/${APP_BINARY}
Restart=always
RestartSec=3
${ENV_LINES}

[Install]
WantedBy=multi-user.target
EOF

  sudo systemctl daemon-reload
  sudo systemctl enable "${SERVICE_NAME}"
fi

echo "==> Restarting service if binary exists"
if [ -x "${DEPLOY_PATH}/${APP_BINARY}" ]; then
  sudo systemctl restart "${SERVICE_NAME}"
  sudo systemctl --no-pager status "${SERVICE_NAME}" -l || true
else
  echo "Binary not found at ${DEPLOY_PATH}/${APP_BINARY}"
  echo "Deploy your binary first, then restart:"
  echo "sudo systemctl restart ${SERVICE_NAME}"
fi

echo "Startup complete."
