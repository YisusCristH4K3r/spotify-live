#!/bin/bash

set -euo pipefail

# Load local environment variables if present.
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

# --- 1. CONFIGURATION (REQUIRED) ---
# SSH details for your VPS
SSH_USER="${SSH_USER:?Missing SSH_USER. Set it in .env}"
SSH_HOST="${SSH_HOST:?Missing SSH_HOST. Set it in .env}"
SSH_PORT="${SSH_PORT:-22}"

# Application details
BINARY_NAME="scraper"
LOCAL_BINARY_PATH="./build/$BINARY_NAME"
REMOTE_DEPLOY_DIR="${REMOTE_DEPLOY_DIR:?Missing REMOTE_DEPLOY_DIR. Set it in .env}"

# Systemd Service details
SERVICE_NAME="spotify-$BINARY_NAME.service"
SERVICE_PATH="/etc/systemd/system/$SERVICE_NAME"
SERVICE_ENV_PATH="/etc/spotify-$BINARY_NAME.env"

# IMPORTANT: This user/group must be created on the remote server
RUNAS_USER="${RUNAS_USER:-spotify-app}"
RUNAS_GROUP="${RUNAS_GROUP:-spotify-app}"

# Application configuration
SP_DC_COOKIE="${SP_DC_COOKIE:?Missing SP_DC_COOKIE. Set it in .env}"
APP_DB="${APP_DB:-database.db}"
# --- END CONFIGURATION ---

# --- 2. LOCAL BUILD CHECK ---
echo "⚙️  Starting local Go build..."
if [ ! -f "$LOCAL_BINARY_PATH" ]; then
    echo "❌ Error: Local binary not found at $LOCAL_BINARY_PATH. Please build it first."
    exit 1
fi
echo "✅ Local build check passed."

# --- 3. DEPLOYMENT FUNCTION ---
echo "🚀 Starting deployment to $SSH_USER@$SSH_HOST:$REMOTE_DEPLOY_DIR"

# Step 3.1: Copy the binary to the remote temporary location
echo "Transferring binary using scp..."
scp -P "$SSH_PORT" "$LOCAL_BINARY_PATH" "$SSH_USER@$SSH_HOST:/tmp/$BINARY_NAME"

# Step 3.2: Execute setup and restart commands remotely via SSH
SSH_COMMANDS=$(cat <<EOF
# --- Remote Setup & Restart ---

set -e

# Create the dedicated user/group and deployment directory if they don't exist
if ! id -u $RUNAS_USER >/dev/null 2>&1; then
    echo "Creating user/group '$RUNAS_USER'..."
    sudo groupadd $RUNAS_GROUP || true
    sudo useradd -r -s /bin/false -g $RUNAS_GROUP $RUNAS_USER || true
fi

# Create deployment directory and set ownership/permissions
sudo mkdir -p $REMOTE_DEPLOY_DIR
sudo chown -R $RUNAS_USER:$RUNAS_GROUP $REMOTE_DEPLOY_DIR
sudo chmod -R 755 $REMOTE_DEPLOY_DIR

# Create/update protected systemd environment file
echo "Writing environment file at $SERVICE_ENV_PATH..."
sudo tee $SERVICE_ENV_PATH > /dev/null <<'ENVEOF'
SP_DC_COOKIE=$SP_DC_COOKIE
APP_DB=$APP_DB
ENVEOF
sudo chown root:$RUNAS_GROUP $SERVICE_ENV_PATH
sudo chmod 640 $SERVICE_ENV_PATH

# Create/update systemd service file
echo "Writing service file at $SERVICE_PATH..."

# NOTE: The inner delimiter 'SERVICE_EOF' is quoted so remote bash won't expand variables.
# The \$ protects the variable from the local bash expansion.
sudo tee $SERVICE_PATH > /dev/null <<'SERVICE_EOF'
[Unit]
Description=Spotify friend list scraping service ($BINARY_NAME)
After=network.target

[Service]
User=$RUNAS_USER
Group=$RUNAS_GROUP
WorkingDirectory=$REMOTE_DEPLOY_DIR
EnvironmentFile=$SERVICE_ENV_PATH

# Systemd natively expands EnvironmentFile variables
ExecStart=$REMOTE_DEPLOY_DIR/$BINARY_NAME --cookie "\${SP_DC_COOKIE}" --db "\${APP_DB}"

Restart=always
RestartSec=5s
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
SERVICE_EOF

# Reload daemon and enable on boot
sudo systemctl daemon-reload
sudo systemctl enable $SERVICE_NAME

# Move the new binary from /tmp to the final deployment directory
echo "Moving new binary into place..."
sudo mv /tmp/$BINARY_NAME $REMOTE_DEPLOY_DIR/$BINARY_NAME
sudo chown $RUNAS_USER:$RUNAS_GROUP $REMOTE_DEPLOY_DIR/$BINARY_NAME
sudo chmod +x $REMOTE_DEPLOY_DIR/$BINARY_NAME

# Restart the service to pick up the new binary/config
echo "Restarting $SERVICE_NAME..."
sudo systemctl restart $SERVICE_NAME

# Check the status of the service
if sudo systemctl is-active --quiet $SERVICE_NAME; then
    echo "✅ Success: $SERVICE_NAME is running with the new binary."
    echo "To check logs, run: journalctl -u $SERVICE_NAME -f"
else
    echo "❌ Error: $SERVICE_NAME failed to start. Check logs immediately."
    sudo systemctl status $SERVICE_NAME --no-pager || true
    exit 1
fi

# --- End Remote Commands ---
EOF
)

# Execute all commands on the remote host
ssh -p "$SSH_PORT" "$SSH_USER@$SSH_HOST" "$SSH_COMMANDS"

echo "✨ Deployment script finished."