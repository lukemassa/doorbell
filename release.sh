#!/bin/bash

set -e
set -o pipefail

ARCH=aarch64-unknown-linux-gnu
TARGET=lmassa@doorbell.local
REMOTE_DIR=/opt/doorbell
REMOTE_BIN=$REMOTE_DIR/bin
LOCAL_BINARY=target/$ARCH/release/doorbell

cross build --target $ARCH  --release

# Upload binary to temp location with scp
scp "$LOCAL_BINARY" "$TARGET:/tmp/doorbell"

# SSH in to move it into place and set permissions
ssh "$TARGET" <<EOF
  sudo mkdir -p $REMOTE_BIN
  sudo mv /tmp/doorbell $REMOTE_BIN/doorbell
  sudo chown root:root $REMOTE_BIN/doorbell
  sudo chmod 755 $REMOTE_BIN/doorbell
EOF
