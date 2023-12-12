#!/bin/bash

# gather input parameters
# -t tag

while getopts ":t:" opt; do
    case $opt in
    t)
        TAG="$OPTARG"
        ;;
    \?)
        echo "Invalid option -$OPTARG" >&2
        ;;
    esac
done

if [ -z "${TAG}" ]; then
    TAG="latest"
fi

echo "TAG = ${TAG}"

required_env_vars=("AUTH_TOKEN_ISS" "AUTH_TOKEN_AUD")

for var in "${required_env_vars[@]}"; do
    if [[ -z "${!var}" ]]; then
        echo "Required environment variable $var is missing"
        exit 1
    fi
done

go build -o actlabs-managed-server ./cmd/actlabs-managed-server

docker build -t actlab.azurecr.io/actlabs-managed-server:${TAG} .

rm actlabs-managed-server

az acr login --name actlab --subscription ACT-CSS-Readiness
docker push actlab.azurecr.io/actlabs-managed-server:${TAG}

docker tag actlab.azurecr.io/actlabs-managed-server:${TAG} ashishvermapu/actlabs-managed-server:${TAG}
docker push ashishvermapu/actlabs-managed-server:${TAG}