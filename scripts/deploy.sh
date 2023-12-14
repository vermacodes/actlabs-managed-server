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
source ../.env

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

required_env_vars=("AUTH_TOKEN_ISS" "AUTH_TOKEN_AUD" "AZURE_CLIENT_ID" "AZURE_CLIENT_SECRET" "AZURE_TENANT_ID")

for var in "${required_env_vars[@]}"; do
  if [[ -z "${!var}" ]]; then
    echo "Required environment variable $var is missing"
    exit 1
  fi
done

# create container app env.
az containerapp env create --name actlabs-managed-server-env \
  --resource-group actlabs-app \
  --subscription ACT-CSS-Readiness \
  --logs-destination none

# create container app
az containerapp create --name actlabs-managed-server \
  --resource-group actlabs-app \
  --subscription ACT-CSS-Readiness \
  --environment actlabs-managed-server-env \
  --allow-insecure false \
  --image actlab.azurecr.io/actlabs-managed-server:${TAG} \
  --ingress 'external' \
  --min-replicas 1 \
  --max-replicas 1 \
  --target-port 80 \
  --env-vars "AUTH_TOKEN_ISS=$AUTH_TOKEN_ISS" "AUTH_TOKEN_AUD=$AUTH_TOKEN_AUD" "AZURE_CLIENT_ID=$AZURE_CLIENT_ID" "AZURE_CLIENT_SECRET=secretref:azure-client-secret" "AZURE_TENANT_ID=$AZURE_TENANT_ID" "LOG_LEVEL=$LOG_LEVEL" \
  --secrets "azure-client-secret=$AZURE_CLIENT_SECRET"
