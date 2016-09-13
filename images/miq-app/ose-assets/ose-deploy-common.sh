#!/bin/bash

[[ -s /etc/default/evm ]] && source /etc/default/evm

# Define persistent dirs paths

DEPLOY_INFO_FILE=${APP_ROOT_PERSISTENT}/.deployment_info

function check_deployment_status {

if [[ -f ${APP_ROOT_PERSISTENT}/config/database.yml && -f ${DEPLOY_INFO_FILE} ]]; then
  echo "== Found existing deployment configuration =="
  echo "== Restoring existing database configuration =="
  mv ${APP_ROOT}/config/database.yml ${APP_ROOT}/config/database.yml.bak
  ln -s ${APP_ROOT_PERSISTENT}/config/database.yml ${APP_ROOT}/config/database.yml
  # Source original deployment env from persistent file
  source ${DEPLOY_INFO_FILE}
  # Obtain currently running env
  CURRENT_APP_VERSION=$(cat ${APP_ROOT}/VERSION)
  CURRENT_SCHEMA_VERSION=$(cd ${APP_ROOT} ; bin/rake db:version | awk '{ print $3 }')
  CURRENT_IMG_VERSION=${IMAGE_VERSION}
  # Check if we have a redeploy with identical versions (exclude master)
  if [[ $APP_VERSION == $CURRENT_APP_VERSION && $APP_VERSION != master ]]; then
    echo "== App version matches original deployment =="
    if [ ${SCHEMA_VERSION} != ${CURRENT_SCHEMA_VERSION} ]; then
       echo "Something seems wrong, db schema version mismatch for the same app version: ${SCHEMA_VERSION} <-> ${CURRENT_SCHEMA_VERSION}"
       exit 1
    fi
  export REDEPLOY=true
  else
  # Assuming different app versions (upgrade)
  echo "== Migrating Database =="
  cd ${APP_ROOT} ; bin/rake db:migrate
  export UPGRADE=true
  fi
else
  echo "Could not find pre-existing configuration, assuming new deployment.."
  export NEW_DEPLOYMENT=true
fi	
}

function write_deployment_info {
# Populate info file based on initial deployment
# IMAGE_VERSION is supplied by docker ENV

APP_VERSION=$(cat ${APP_ROOT}/VERSION)
SCHEMA_VERSION=$(cd ${APP_ROOT} ; bin/rake db:version | awk '{ print $3 }')

if [[ -z $APP_VERSION || -z $SCHEMA_VERSION || -z $IMAGE_VERSION ]]; then
  echo "${DEPLOY_INFO_FILE} is incomplete, one or more required variables are undefined"
  exit 1
else
  echo "APP_VERSION=${APP_VERSION}" > ${DEPLOY_INFO_FILE}
  echo "SCHEMA_VERSION=${SCHEMA_VERSION}" >> ${DEPLOY_INFO_FILE}
  echo "IMG_VERSION=${IMAGE_VERSION}" >> ${DEPLOY_INFO_FILE}
fi
}

function setup_persistent_data {

APP_ROOT_PERSISTENT="/var/www/miq/vmdb.test/persistent"
CONTAINER_INPUT_FILE="/container.data.persist"
PERSIST_INPUT_FILE="$APP_ROOT_PERSISTENT/container.data.persist"

 [ ! -f "${PERSIST_INPUT_FILE}" ] && cp -a ${CONTAINER_INPUT_FILE} ${APP_ROOT_PERSISTENT}

while read -r file
do
    # Sanity checks
    [[ $file = \#* ]] && continue
    [[ $file == / ]] && continue
    [[ ! -e $file ]] && echo "$file does not exist, skipping" && continue
    [[ -h $file ]] && echo "$file symblink is already in place, skipping" && continue
    # Obtain dirname and filename from source file
    DIR=$(dirname $file)
    FILENAME=$(basename $file)
    # Create directory structure under persistent volume if not present
    [[ ! -d ${APP_ROOT_PERSISTENT}/${DIR} ]] && mkdir -p ${APP_ROOT_PERSISTENT}/${DIR}
    # Copy data into persistent volume
    cp -a $file ${APP_ROOT_PERSISTENT}${DIR}/$FILENAME
    # Check if we are working with a directory, backup
    [[ -d $file ]] && mv ${file} ${file}~
    # Place symblink back to persistent volume
    ln --backup -sn ${APP_ROOT_PERSISTENT}${DIR}/$FILENAME $file
done < "${PERSIST_INPUT_FILE}"

}
