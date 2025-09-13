#!/bin/bash

set -e
set -o pipefail

function usage() {
    echo "Usage: [--restart]"
    exit 1
}

restart=false
while [[ "$#" -gt 0 ]]
do
    arg="$1"
    shift
    if [[ "$arg" == "--restart" ]]
    then
       restart=true
    else
       usage
    fi
done

GOOS=linux
GOARCH=arm64
TARGET=lmassa@doorbell.local
REMOTE_DIR=/opt/doorbell
REMOTE_BIN=$REMOTE_DIR/bin
LOCAL_BINARY=main
LOCAL_CONF_FILE="conf/config.yaml"

go run cmd/main.go validate
go test ./...

GOOS=$GOOS GOARCH=$GOARCH go build cmd/main.go

# Upload binary and conf files to temp location with scp
scp "$LOCAL_BINARY" "$TARGET:/tmp/doorbell"
scp "$LOCAL_CONF_FILE" "$TARGET:/tmp/doorbell-config.yaml"

finalcmd="echo 'Exiting without restarting'"
if $restart
then
  finalcmd="sudo service doorbell restart"
fi

# SSH in to move it into place and set permissions
ssh "$TARGET" <<EOF
  sudo mkdir -p $REMOTE_BIN
  sudo mv /tmp/doorbell $REMOTE_BIN/doorbell
  sudo mv /tmp/doorbell-config.yaml $REMOTE_DIR/conf/config.yaml
  sudo chown root:root $REMOTE_BIN/doorbell
  sudo chmod 755 $REMOTE_BIN/doorbell
  $finalcmd
EOF
