#!/bin/bash

# Generate a new topic with the given age.key

function usage() {
    echo "Usage: prefix"
    exit 1
}

if [[ "$1" == "" || "$1" == -* ]]
then
    usage
fi
prefix="$1"

if ! [[ "$prefix" =~ ^[0-9a-z_]+$ ]]
then
    echo "Unexpected prefix: $prefix"
    exit 1
fi

ageKeyFile=$(cat conf/config.yaml | yq .ageKeyFile)
agePub=$(age-keygen -y $ageKeyFile)

uuid=$(uuidgen)
topic="$prefix-$uuid"

echo "Unencrypted Topic $topic"
echo "Encrypted Topic   $(echo $topic | age -r $agePub | base64)"
