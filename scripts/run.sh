#!/bin/bash

# gather input parameters
# if flag -d is set the set LOG_LEVEL to -4 else to 0

while getopts ":d" opt; do
    case $opt in
    d)
        LOG_LEVEL="-4"
        ;;
    \?)
        echo "Invalid option -$OPTARG" >&2
        ;;
    esac
done

if [ -z "${LOG_LEVEL}" ]; then
    LOG_LEVEL="0"
fi

echo "LOG_LEVEL = ${LOG_LEVEL}"

export ROOT_DIR=$(pwd)

required_env_vars=("AUTH_TOKEN_ISS" "AUTH_TOKEN_AUD")

for var in "${required_env_vars[@]}"; do
    if [[ -z "${!var}" ]]; then
        echo "Required environment variable $var is missing"
        exit 1
    fi
done

rm ./tmp/main

go build -o ./tmp/main ./cmd/actlabs-managed-server

export LOG_LEVEL="${LOG_LEVEL}" && export PORT="8883"