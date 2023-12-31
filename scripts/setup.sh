#!/bin/bash

ACTLABS_APP_ID="9ae3061c-c01c-4d5d-88d6-ca2cb5c66c6e"
RESOURCE_GROUP="repro-project"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to log messages
err() {
    echo -e "${RED}[$(date +'%Y-%m-%dT%H:%M:%S%z')]: ERROR - $* ${NC}" >&1
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%dT%H:%M:%S%z')]: WARNING - $* ${NC}" >&1
}

log() {
    echo -e "[$(date +'%Y-%m-%dT%H:%M:%S%z')]: INFO - $*" >&1
}

ok() {
    echo -e "${GREEN}[$(date +'%Y-%m-%dT%H:%M:%S%z')]: SUCCESS - $* ${NC}" >&1
}

gap() {
    echo >&1
}

# Function that sleeps for a specified number of seconds
function sleep_with_progress() {
    local TOTAL_SECONDS=$1
    local MESSAGE=$2 # optional; if provided, will be printed before the countdown

    # Check if TOTAL_SECONDS is a positive integer
    if ! [[ "$TOTAL_SECONDS" =~ ^[0-9]+$ ]]; then
        log "Error: Invalid number of seconds"
        return 1
    fi

    if [[ -n "${MESSAGE}" ]]; then
        ok "${MESSAGE}"
    fi

    local MINUTES=$((TOTAL_SECONDS / 60))
    local SECONDS_REMAINING=$((TOTAL_SECONDS % 60))
    log "Sleeping for ${MINUTES} minutes and ${SECONDS_REMAINING} seconds..."

    while [[ ${TOTAL_SECONDS} -gt 0 ]]; do
        MINUTES=$((TOTAL_SECONDS / 60))
        SECONDS_REMAINING=$((TOTAL_SECONDS % 60))
        log "${MINUTES} minutes and ${SECONDS_REMAINING} seconds remaining"
        sleep 10
        TOTAL_SECONDS=$((TOTAL_SECONDS - 10))
        log "Seconds remaining: ${TOTAL_SECONDS}"
    done
}

# Function to handle errors
handle_error() {
    err "$1"
    exit 1
}

# Function to get upn of the logged in user
get_upn() {
    UPN=$(az ad signed-in-user show --query userPrincipalName -o tsv)
    log "UPN: $UPN"

    # drop the domain name from the upn
    if [[ "${UPN}" == *"fdpo.onmicrosoft.com"* ]]; then
        # USER_ALIAS=${UPN%%_*}
        handle_error "We currently do not support Microsoft Non-Prod Tenant. Please reach out to the team for support."
    else
        USER_ALIAS=${UPN%%@*}
    fi
    log "USER_ALIAS: $USER_ALIAS"

    USER_ALIAS_FOR_SA=${USER_ALIAS#v-}
    log "USER_ALIAS_FOR_SA: $USER_ALIAS_FOR_SA"
}

# Function to get the subscription id
get_subscription_id() {
    SUBSCRIPTION_ID=$(az account show --query id -o tsv)
    log "SUBSCRIPTION_ID: $SUBSCRIPTION_ID"
}

# Function to check if a resource group exists
# If the resource group doesn't exist, create one
function create_resource_group() {
    # Check if the resource group exists
    RG_EXISTS=$(az group exists --name "${RESOURCE_GROUP}" --output tsv)

    if [[ "${RG_EXISTS}" == "true" ]]; then
        log "resource group ${RESOURCE_GROUP} already exists"
        return 0
    else
        log "creating resource group with name ${RESOURCE_GROUP}"

        # Ask the user for a location if one wasn't provided
        if [[ -z "${LOCATION}" ]]; then
            gap
            LOCATION=$(az account list-locations --query "[?metadata.regionType!='Logical' && metadata.physicalLocation!=null].name" -o tsv)
            echo "Please select a location (azure region) from the list below:"
            select LOCATION in ${LOCATION}; do
                if [[ -n "${LOCATION}" ]]; then
                    break
                else
                    echo "Invalid selection, please try again"
                fi
            done
        fi

        # Create the resource group
        az group create --name "${RESOURCE_GROUP}" --location "${LOCATION}"
        if [ $? -ne 0 ]; then
            err "failed to create resource group ${RESOURCE_GROUP}"
            return 1
        else
            log "resource group ${RESOURCE_GROUP} created"
        fi
    fi

    return 0
}

# Function to check if a storage account exists in the resource group
# If the storage account doesn't exist, create one with a random name
function create_storage_account() {
    # Check if the storage account exists in the resource group
    SA_EXISTS=$(az storage account list --resource-group "${RESOURCE_GROUP}" --query "[].name" -o tsv)

    if [[ -n "${SA_EXISTS}" ]]; then
        log "storage account already exists with name ${SA_EXISTS}"
        STORAGE_ACCOUNT_NAME="$SA_EXISTS"
    else
        # Generate a random name for the storage account
        RANDOM_NAME=$(openssl rand -hex 4)
        STORAGE_ACCOUNT_NAME="${USER_ALIAS_FOR_SA}sa${RANDOM_NAME}"

        log "creating storage account with name ${STORAGE_ACCOUNT_NAME} in resource group ${RESOURCE_GROUP}"
        # Create the storage account
        az storage account create --name "${STORAGE_ACCOUNT_NAME}" --resource-group "${RESOURCE_GROUP}" --sku Standard_LRS
        if [ $? -ne 0 ]; then
            err "failed to create storage account ${STORAGE_ACCOUNT_NAME}"
            return 1
        else
            log "storage account ${STORAGE_ACCOUNT_NAME} created"
        fi
    fi

    # get storage account key
    STORAGE_ACCOUNT_KEY=$(az storage account keys list --resource-group "${RESOURCE_GROUP}" --account-name "${STORAGE_ACCOUNT_NAME}" --query "[0].value" -o tsv)

    # check if a blob container named 'tfstate' exists in the storage account
    # if not create one
    log "checking if blob container tfstate exists in storage account ${STORAGE_ACCOUNT_NAME}"
    CONTAINER_EXISTS=$(az storage container exists --name "tfstate" --account-name "${STORAGE_ACCOUNT_NAME}" --account-key "${STORAGE_ACCOUNT_KEY}" --query "exists" -o tsv)
    if [[ "${CONTAINER_EXISTS}" == "true" ]]; then
        log "Blob container tfstate already exists in storage account ${STORAGE_ACCOUNT_NAME}"
    else
        log "Blob container tfstate does not exist in storage account ${STORAGE_ACCOUNT_NAME}, creating"
        az storage container create --name "tfstate" --account-name "${STORAGE_ACCOUNT_NAME}"
        if [ $? -ne 0 ]; then
            err "Failed to create blob container tfstate in storage account ${STORAGE_ACCOUNT_NAME}"
            return 1
        else
            log "Blob container tfstate created in storage account ${STORAGE_ACCOUNT_NAME}"
        fi
    fi

    # check if a blob container named 'labs' exists in the storage account
    # if not create one
    log "checking if blob container labs exists in storage account ${STORAGE_ACCOUNT_NAME}"
    CONTAINER_EXISTS=$(az storage container exists --name "labs" --account-name "${STORAGE_ACCOUNT_NAME}" --account-key "${STORAGE_ACCOUNT_KEY}" --query "exists" -o tsv)
    if [[ "${CONTAINER_EXISTS}" == "true" ]]; then
        log "Blob container labs already exists in storage account ${STORAGE_ACCOUNT_NAME}"
    else
        log "Blob container labs does not exist in storage account ${STORAGE_ACCOUNT_NAME}, creating"
        az storage container create --name "labs" --account-name "${STORAGE_ACCOUNT_NAME}"
        if [ $? -ne 0 ]; then
            err "Failed to create blob container labs in storage account ${STORAGE_ACCOUNT_NAME}"
            return 1
        else
            log "Blob container labs created in storage account ${STORAGE_ACCOUNT_NAME}"
        fi
    fi

    return 0
}

# Function to check if a managed identity exists in the resource group
# If the managed identity doesn't exist, create one
function create_managed_identity() {
    # Check if the managed identity exists in the resource group
    MI_EXISTS=$(az identity list --resource-group "${RESOURCE_GROUP}" --query "[?name=='${USER_ALIAS}-msi'].name" -o tsv)

    if [[ -n "${MI_EXISTS}" ]]; then
        log "managed identity already exists with name ${MI_EXISTS}"
        MANAGED_IDENTITY_NAME="$MI_EXISTS"
    else
        log "creating managed identity with name ${USER_ALIAS}-msi in resource group ${RESOURCE_GROUP}"
        # Create the managed identity
        az identity create --name "${USER_ALIAS}-msi" --resource-group "${RESOURCE_GROUP}"
        if [ $? -ne 0 ]; then
            err "failed to create managed identity ${USER_ALIAS}-msi"
            return 1
        else
            log "managed identity ${USER_ALIAS}-msi created"
            sleep_with_progress 120 "Waiting for the managed identity to be created and synced to Azure AD"
        fi
    fi

    # Get the managed identity client id
    MANAGED_IDENTITY_CLIENT_ID=$(az identity show --name "${USER_ALIAS}-msi" --resource-group "${RESOURCE_GROUP}" --query clientId -o tsv)

    return 0
}

# Function to check if managed identity is 'Contributor' on the subscription
# If not, assign the managed identity the 'Contributor' role on the subscription
function assign_contributor_role() {
    # Check if the managed identity is 'Contributor' on the subscription
    MI_ROLE=$(az role assignment list --assignee "${MANAGED_IDENTITY_CLIENT_ID}" --scope "/subscriptions/${SUBSCRIPTION_ID}" --query "[?roleDefinitionName=='Contributor'].roleDefinitionName" -o tsv)

    if [[ -n "${MI_ROLE}" ]]; then
        log "managed identity ${USER_ALIAS}-msi is already 'Contributor' on the subscription"
    else
        log "assigning managed identity ${USER_ALIAS}-msi 'Contributor' role on the subscription"
        # Assign the managed identity the 'Contributor' role on the subscription
        az role assignment create --assignee "${MANAGED_IDENTITY_CLIENT_ID}" --role Contributor
        if [ $? -ne 0 ]; then
            err "failed to assign managed identity ${USER_ALIAS}-msi 'Contributor' role on the subscription"
            return 1
        else
            log "managed identity ${USER_ALIAS}-msi assigned 'Contributor' role on the subscription"
        fi
    fi

    return 0
}

# Function to check if managed identity is 'User Access Administrator' on the subscription
# If not, assign the managed identity the 'User Access Administrator' role on the subscription
function assign_user_access_administrator_role() {
    # Check if the managed identity is 'User Access Administrator' on the subscription
    MI_ROLE=$(az role assignment list --assignee "${MANAGED_IDENTITY_CLIENT_ID}" --scope "/subscriptions/${SUBSCRIPTION_ID}" --query "[?roleDefinitionName=='User Access Administrator'].roleDefinitionName" -o tsv)

    if [[ -n "${MI_ROLE}" ]]; then
        log "managed identity ${USER_ALIAS}-msi is already 'User Access Administrator' on the subscription"
    else
        log "assigning managed identity ${USER_ALIAS}-msi 'User Access Administrator' role on the subscription"
        # Assign the managed identity the 'User Access Administrator' role on the subscription
        az role assignment create --assignee "${MANAGED_IDENTITY_CLIENT_ID}" --role "User Access Administrator"
        if [ $? -ne 0 ]; then
            err "failed to assign managed identity ${USER_ALIAS}-msi 'User Access Administrator' role on the subscription"
            return 1
        else
            log "managed identity ${USER_ALIAS}-msi assigned 'User Access Administrator' role on the subscription"
        fi
    fi

    return 0
}

# Function to assign actlabs as a contributor on the resource group
function assign_actlabs_contributor_role() {
    # Check if actlabs is a contributor on the resource group
    ACTLABS_ROLE=$(az role assignment list --assignee "${ACTLABS_APP_ID}" --scope "/subscriptions/${SUBSCRIPTION_ID}/resourceGroups/${RESOURCE_GROUP}" --query "[?roleDefinitionName=='Contributor'].roleDefinitionName" -o tsv)

    if [[ -n "${ACTLABS_ROLE}" ]]; then
        log "actlabs is already 'Contributor' on the resource group"
    else
        log "assigning actlabs 'Contributor' role on the resource group"
        # Assign actlabs the 'Contributor' role on the resource group
        az role assignment create --assignee "${ACTLABS_APP_ID}" --role Contributor --resource-group "${RESOURCE_GROUP}"
        if [ $? -ne 0 ]; then
            err "failed to assign actlabs 'Contributor' role on the resource group"
            return 1
        else
            log "actlabs assigned 'Contributor' role on the resource group"
        fi
    fi

    return 0
}

# Function to assign actlabs as a reader on the subscription
function assign_actlabs_reader_role() {
    # Check if actlabs is a reader on the subscription
    ACTLABS_ROLE=$(az role assignment list --assignee "${ACTLABS_APP_ID}" --scope "/subscriptions/${SUBSCRIPTION_ID}" --query "[?roleDefinitionName=='Reader'].roleDefinitionName" -o tsv)

    if [[ -n "${ACTLABS_ROLE}" ]]; then
        log "actlabs is already 'Reader' on the subscription"
    else
        log "assigning actlabs 'Reader' role on the subscription"
        # Assign actlabs the 'Reader' role on the subscription
        az role assignment create --assignee "${ACTLABS_APP_ID}" --role Reader --subscription "${SUBSCRIPTION_ID}"
        if [ $? -ne 0 ]; then
            err "failed to assign actlabs 'Reader' role on the subscription"
            return 1
        else
            log "actlabs assigned 'Reader' role on the subscription"
        fi
    fi

    return 0
}

# Function to print the instructions to the user
function print_next_steps() {
    ok "Deployment complete!"
    ok "Next steps: Go back to UI and register your subscription with the lab"
    ok "Your subscription id is: ${SUBSCRIPTION_ID}"
}

# Call the functions
get_upn
get_subscription_id
create_resource_group
create_storage_account
create_managed_identity
assign_contributor_role
assign_user_access_administrator_role
assign_actlabs_contributor_role
assign_actlabs_reader_role
print_next_steps
