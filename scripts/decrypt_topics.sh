#!/bin/bash

set -e
set -o pipefail

function run_yq() {
    local query="$1"
    cat conf/config.yaml | yq "$query"
}

ageKeyFile=$(run_yq .ageKeyFile)

# TODO: Pull out keys in particular
for unit in $(run_yq '.units|keys' | awk '{print $2}')
do
    echo "Topics for $unit"
    num_on_press=$(run_yq ".units.$unit.on_press|length")
    for ((i=0; i<$num_on_press; i++)); do
        if [[ "$(run_yq ".units.$unit.on_press[$i] | has(\"ntfy\")")" != "true" ]]
        then
            continue
        fi
        run_yq ".units.$unit.on_press[$i].ntfy.encryptedTopic" | base64 -d | age --decrypt -i $ageKeyFile
    done
    #echo $unit
done
#cat conf/config.yaml | yq '.units[].on_press[0].ntfy.encryptedTopic' | base64 -d | age --decrypt -i $ageKeyFile
