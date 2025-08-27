#!/bin/bash

set -euo pipefail

function usage() {
	echo "Usage: duration"
	exit 1
}

DURATION=60
if [[ "$#" -gt 0 ]]
then
	if [[ "$1" = -* ]]
	then
		usage
	fi
	DURATION="$1"
fi

MQTT_TOPIC="zigbee2mqtt/bridge/request/permit_join"
MQTT_PAYLOAD="{\"time\": $DURATION}"

echo "Enabling Zigbee pairing for $DURATION seconds..."
mosquitto_pub -t "$MQTT_TOPIC" -m "$MQTT_PAYLOAD"

./logs.sh
