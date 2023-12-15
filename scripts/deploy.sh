#!/bin/bash

# gather input parameters
# -t "tag" for image tag
# -d for debug

while getopts ":t:d" opt; do
    case $opt in
    t)
        TAG=$OPTARG
        ;;
    d)
        DEBUG=true
        ;;
    \?)
        echo "Invalid option: -$OPTARG" >&2
        ;;
    esac
done

source .env

echo "TAG = ${TAG}"
echo "DEBUG = ${DEBUG}"

if [ -z "${TAG}" ]; then
    TAG="latest"
fi

if [ ${DEBUG} ]; then
    export LOG_LEVEL="-4"
else
    export LOG_LEVEL="0"
fi

required_env_vars=("AUTH_TOKEN_ISS" "AUTH_TOKEN_AUD" "SERVER_MANAGER_CLIENT_ID" "TENANT_ID")

for var in "${required_env_vars[@]}"; do
  if [[ -z "${!var}" ]]; then
    echo "Required environment variable $var is missing"
    exit 1
  fi
done

# create container group
az container create \
  --resource-group actlabs-app \
  --subscription ACT-CSS-Readiness \
  --file ./deploy/deploy.yaml

