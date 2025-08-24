#!/bin/bash

set -e
set -o pipefail

GOOS=linux
GOARCH=arm64
TARGET=lmassa@doorbell.local
REMOTE_DIR=/opt/doorbell
REMOTE_BIN=$REMOTE_DIR/bin
LOCAL_BINARY=main

GOOS=$GOOS GOARCH=$GOARCH go build cmd/main.go

# Upload binary to temp location with scp
scp "$LOCAL_BINARY" "$TARGET:/tmp/doorbell"

# SSH in to move it into place and set permissions
ssh "$TARGET" <<EOF
  sudo mkdir -p $REMOTE_BIN
  sudo mv /tmp/doorbell $REMOTE_BIN/doorbell
  sudo chown root:root $REMOTE_BIN/doorbell
  sudo chmod 755 $REMOTE_BIN/doorbell
  sudo service doorbell restart
EOF
