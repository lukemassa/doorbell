#!/bin/sh

set -eu

RESOLVER=$(awk '/^nameserver/ {print $2; exit}' /etc/resolv.conf || true)

if [ -z "$RESOLVER" ]; then
    echo "No resolver found in /etc/resolv.conf"
    exit 1
fi

echo "Waiting for DNS resolver at $RESOLVER:53 ..."
for i in $(seq 1 5); do   # 5 tries = ~10 seconds at 2s each
    if nc -z "$RESOLVER" 53 >/dev/null 2>&1; then
        echo "Resolver reachable"
        exit 0
    fi
    sleep 2
done

echo "Timeout: resolver not reachable after 10s"
exit 1
