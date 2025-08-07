#!/bin/bash
set -euo pipefail

# --- CONFIGURABLE VERSION ---
NODE_VERSION="v24.4.1"
APP_DIR=/opt/zigbee2mqtt

# --- Create user if not exists ---
if ! id zigbee &>/dev/null; then
  sudo useradd -r -m -d /home/zigbee -s /bin/bash zigbee
  sudo usermod -a -G dialout zigbee
fi

# --- Install NVM and Node (if not already present) ---
if [ ! -d /home/zigbee/.nvm ]; then
  sudo -u zigbee bash <<EOF
    set -euo pipefail
    export NVM_DIR="/home/zigbee/.nvm"
    git clone https://github.com/nvm-sh/nvm.git "\$NVM_DIR"
    cd "\$NVM_DIR"
    git checkout \$(git describe --abbrev=0 --tags)
    . "\$NVM_DIR/nvm.sh"
    nvm install $NODE_VERSION
    nvm alias default $NODE_VERSION
EOF
fi

# --- Setup Zigbee2MQTT repo ---
if [ ! -d $APP_DIR ]; then
  git clone --depth 1 https://github.com/Koenkk/zigbee2mqtt.git $APP_DIR
  chown -R zigbee:zigbee $APP_DIR
fi

# Copying configuration.yaml file
if [ ! -f ${APP_DIR}/data/configuration.yaml ]; then
  cp configuration.yaml ${APP_DIR}/data/configuration.yaml
fi

# --- Install dependencies 
sudo -u zigbee bash <<EOF
  set -euo pipefail
  export NVM_DIR="/home/zigbee/.nvm"
  . "\$NVM_DIR/nvm.sh"
  nvm use $NODE_VERSION
  cd $APP_DIR
  corepack enable
  corepack prepare pnpm@latest --activate
  pnpm install

EOF

# --- Install Mosquitto MQTT broker ---
apt-get update
apt-get install -y mosquitto mosquitto-clients

systemctl enable mosquitto
systemctl start mosquitto

# --- Create systemd unit file ---
cat zigbee2mqtt.service.tmpl | sed "s|\${NODE_VERSION}|$NODE_VERSION|g" > /etc/systemd/system/zigbee2mqtt.service
echo "Wrote systemd unit with Node version: $NODE_VERSION"

# --- Reload and start ---
systemctl daemon-reexec
systemctl daemon-reload
systemctl enable zigbee2mqtt
systemctl start zigbee2mqtt
