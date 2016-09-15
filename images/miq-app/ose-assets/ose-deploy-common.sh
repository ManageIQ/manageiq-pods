#!/bin/bash

[[ -s /etc/default/evm ]] && source /etc/default/evm

# This file is created by the write_deployment_info during initial deployment
PV_DEPLOY_INFO_FILE=${APP_ROOT_PERSISTENT}/.deployment_info

# This file is supplied by the app docker image with default files/dirs to persist on PV
CONTAINER_DATA_PERSIST_FILE="/container.data.persist"

# Copy of CONTAINER_DATA_PERSIST_FILE that will be stored on PV and can be customized by users to add more files/dirs
PV_DATA_PERSIST_FILE="$APP_ROOT_PERSISTENT/container.data.persist"

# VMDB app_root directory inside persistent volume mount
APP_ROOT_PERSISTENT_VMDB=${APP_ROOT_PERSISTENT}/var/www/miq/vmdb

function check_deployment_status {
# Description
# Inspect PV for previous deployments, if a DB a config is present, restore
# Source previous deployment info file from PV and compare data with current environment
# Evaluate conditions and decide a target deployment type: redeploy,upgrade or new

echo "== Checking deployment status =="

if [[ -f ${APP_ROOT_PERSISTENT_VMDB}/config/database.yml && -f ${PV_DEPLOY_INFO_FILE} ]]; then
  echo "== Found existing deployment configuration =="
  echo "== Restoring existing database configuration =="
  ln --backup -sn ${APP_ROOT_PERSISTENT_VMDB}/config/database.yml ${APP_ROOT}/config/database.yml
  # Source original deployment info variables from PV
  source ${PV_DEPLOY_INFO_FILE}
  # Obtain current running environment
  APP_VERSION=$(cat ${APP_ROOT}/VERSION)
  SCHEMA_VERSION=$(cd ${APP_ROOT} && env RAILS_USE_MEMORY_STORE=true bin/rake db:version | awk '{ print $3 }')
  IMG_VERSION=${IMAGE_VERSION}
  # Check if we have identical EVM versions (exclude master builds)
  if [[ $APP_VERSION == $PV_APP_VERSION && $APP_VERSION != master ]]; then
    echo "== App version matches original deployment =="
  # Check if we have same schema version for same EVM version
    if [ "${SCHEMA_VERSION}" != "${PV_SCHEMA_VERSION}" ]; then
       echo "Something seems wrong, db schema version mismatch for the same app version: ${PV_SCHEMA_VERSION} <-> ${SCHEMA_VERSION}"
       exit 1
    fi
  # Assuming redeployment (same APP_VERSION)
  export REDEPLOY=true
  else
  # Assuming upgrade (different APP_VERSION)
  export UPGRADE=true
  fi
else
  echo "No pre-existing EVM configuration found on PV"
  export NEW_DEPLOYMENT=true
fi

}

function write_deployment_info {
# Description
# Populate info file based on initial deployment and store on PV
# Output in bash format to be easily sourced
# IMAGE_VERSION is supplied by docker environment

[ -f "${PV_DEPLOY_INFO_FILE}" ] && { echo "Something seems wrong, ${PV_DEPLOY_INFO_FILE} already exists on a new deployment"; exit 1; }

DEPLOYMENT_DATE=$(date +"%F_%T")
APP_VERSION=$(cat ${APP_ROOT}/VERSION)
SCHEMA_VERSION=$(cd ${APP_ROOT} && bin/rake db:version | awk '{ print $3 }')

if [[ -z $APP_VERSION || -z $SCHEMA_VERSION || -z $IMAGE_VERSION ]]; then
  echo "${PV_DEPLOY_INFO_FILE} is incomplete, one or more required variables are undefined"
  exit 1
else
  echo "PV_APP_VERSION=${APP_VERSION}" > ${PV_DEPLOY_INFO_FILE}
  echo "PV_SCHEMA_VERSION=${SCHEMA_VERSION}" >> ${PV_DEPLOY_INFO_FILE}
  echo "PV_IMG_VERSION=${IMAGE_VERSION}" >> ${PV_DEPLOY_INFO_FILE}
  echo "PV_DEPLOYMENT_DATE=${DEPLOYMENT_DATE}" >> ${PV_DEPLOY_INFO_FILE}
fi

}

function prepare_svc_vars {

# Description
# Prepare service host variables for use in other functions
# *_SERVICE_NAME variables are supplied via OpenShift template parameters (i.e DATABASE_SERVICE_NAME)
# *_SERVICE_HOST variables are auto-created by OpenShift and contain the IP address exposed by service pod
# Upcase values of SERVICE_NAME, assemble proper SVC_HOST and export vars for use

export DATABASE_SVC_HOST="$(echo $DATABASE_SERVICE_NAME | tr '[:lower:]' '[:upper:]')_SERVICE_HOST"
export MEMCACHED_SVC_HOST="$(echo $MEMCACHED_SERVICE_NAME | tr '[:lower:]' '[:upper:]')_SERVICE_HOST"

}

function setup_memcached {
# Description
# Replace memcached host in EVM configuration to use assigned service pod IP

echo "== Applying memcached config =="
sed -i~ -E "s/:memcache_server:.*/:memcache_server: ${!MEMCACHED_SVC_HOST}:11211/gi" ${APP_ROOT}/config/settings.yml

}

function setup_persistent_data {
# Description
# Process container.data.persist which contains the desired files/dirs to store on PV
# Copy listed files/dirs to PV, make backups and deploy symblinks pointing to PV

[ ! -f "${PV_DATA_PERSIST_FILE}" ] && cp -a ${CONTAINER_DATA_PERSIST_FILE} ${APP_ROOT_PERSISTENT}

echo "== Initializing persistent data =="

while read -r FILE
do
    # Sanity checks
    [[ ${FILE} = \#* ]] && continue
    [[ ${FILE} == / ]] && continue
    [[ ! -e ${FILE} ]] && echo "${FILE} does not exist, skipping" && continue
    [[ -h ${FILE} ]] && echo "${FILE} symblink is already in place, skipping" && continue
    # Obtain dirname and filename from source file
    DIR=$(dirname ${FILE})
    FILENAME=$(basename ${FILE})
    # Create directory structure under persistent volume if not present
    [[ ! -d ${APP_ROOT_PERSISTENT}/${DIR} ]] && mkdir -p ${APP_ROOT_PERSISTENT}/${DIR}
    # Copy supplied files/dirs into persistent volume
    cp -a ${FILE} ${APP_ROOT_PERSISTENT}${DIR}/${FILENAME}
    # Check if we are working with a directory, backup
    [[ -d ${FILE} ]] && mv ${FILE} ${FILE}~
    # Place symblink back to persistent volume
    ln --backup -sn ${APP_ROOT_PERSISTENT}${DIR}/${FILENAME} ${FILE}
done < "${PV_DATA_PERSIST_FILE}"

}

function restore_persistent_data {

# Description
# Process container.data.persist which contains the desired files/dirs to restore from PV
# Check if file/dir exits on PV, redeploy symblinks on ${APP_ROOT} pointing to PV

# Ensure PV_DATA_PERSIST_FILE is present, it should be if setup_persistent_data was executed
[ ! -f "${PV_DATA_PERSIST_FILE}" ] && cp -a ${CONTAINER_DATA_PERSIST_FILE} ${APP_ROOT_PERSISTENT}

echo "== Restoring persistent data =="

while read -r FILE
do
    # Sanity checks
    [[ ${FILE} = \#* ]] && continue
    [[ ${FILE} == / ]] && continue
    [[ ! -e ${APP_ROOT_PERSISTENT}$FILE ]] && echo "${FILE} does not exist on PV, skipping" && continue
    [[ -h ${FILE} ]] && echo "${FILE} symblink is already in place, skipping" && continue
    # Obtain dirname and filename from source file
    DIR=$(dirname ${FILE})
    FILENAME=$(basename ${FILE})
    # Check if we are working with a directory, backup
    [[ -d ${FILE} ]] && mv ${FILE} ${FILE}~
    # Place symblink back to persistent volume
    ln --backup -sn ${APP_ROOT_PERSISTENT}${DIR}/${FILENAME} ${FILE}
done < "${PV_DATA_PERSIST_FILE}"

}
