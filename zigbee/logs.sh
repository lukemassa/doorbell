#!/bin/bash bash

logfile=$(find /opt/zigbee2mqtt/data/log -type f -printf '%T@ %p\n' | sort -k1,1nr | head -n 1 | awk '{print $2}')
tail -f $logfile
