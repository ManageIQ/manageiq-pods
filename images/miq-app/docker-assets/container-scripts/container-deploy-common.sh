#!/bin/bash

[[ -s /etc/default/evm ]] && source /etc/default/evm

# This file is created by the write_deployment_info during initial deployment
PV_DEPLOY_INFO_FILE="${APP_ROOT_PERSISTENT}/.deployment_info"

# This directory is used to store server specific data to be persisted
PV_CONTAINER_DATA_DIR="${APP_ROOT_PERSISTENT}/container-data"

# This directory is used to store server specific container deployment data (logs,backups,etc)
PV_CONTAINER_DEPLOY_DIR="${APP_ROOT_PERSISTENT}/container-deploy"

# This directory is used to store server specific initialization logfiles on PV
PV_LOG_DIR="${PV_CONTAINER_DEPLOY_DIR}/log"

# Directory used to backup server specific PV data before performing an upgrade
PV_BACKUP_DIR="${PV_CONTAINER_DEPLOY_DIR}/backup"

# This directory is used to store shared region application data to be persisted (database.yml, keys, etc)
PV_CONTAINER_DATA_REGION_DIR="${APP_ROOT_PERSISTENT_REGION}/container-data"

# This file is supplied by the app docker image with default files/dirs to persist on PV
CONTAINER_DATA_PERSIST_FILE="/container.data.persist"

# Copy of CONTAINER_DATA_PERSIST_FILE that will be stored on PV and can be customized by users to add more files/dirs
PV_DATA_PERSIST_FILE="${APP_ROOT_PERSISTENT}/container.data.persist"

# Set log timestamp for running instance
PV_LOG_TIMESTAMP="$(date +%s)"

# VMDB shared REGION app_root directory on PV
PV_REGION_VMDB="${PV_CONTAINER_DATA_REGION_DIR}/var/www/miq/vmdb"

function check_deployment_status() {
# Description
# Inspect PV for previous deployments, if a DB a config is present, restore
# Source previous deployment info file from PV and compare data with current environment
# Evaluate conditions and decide a target deployment type: redeploy,upgrade or new

echo "== Checking deployment status =="

if [[ -f ${PV_REGION_VMDB}/config/database.yml && -f ${PV_DEPLOY_INFO_FILE} ]]; then
  echo "== Found existing deployment configuration =="
  echo "== Restoring existing database configuration =="
  ln --backup -sn ${PV_REGION_VMDB}/config/database.yml ${APP_ROOT}/config/database.yml
  [[ ! -f ${PV_REGION_VMDB}/certs/v2_key ]] && echo "ERROR: Could not find ${PV_REGION_VMDB}/certs/v2_key on upgrade/redeploy case, aborting.." && exit 1
  ln --backup -sn ${PV_REGION_VMDB}/certs/v2_key ${APP_ROOT}/certs/v2_key
  # Source original deployment info variables from PV
  source ${PV_DEPLOY_INFO_FILE}
  # Obtain current running environment
  APP_VERSION="$(cat ${APP_ROOT}/VERSION)"
  SCHEMA_VERSION="$(cd ${APP_ROOT} && RAILS_USE_MEMORY_STORE=true bin/rake db:version | awk '{ print $3 }')"
  # Check if we have identical EVM versions (exclude master builds)
  if [[ ${APP_VERSION} == ${PV_APP_VERSION} && ${APP_VERSION} != master ]]; then
    echo "== App version matches original deployment =="
  # Check if we have same schema version for same EVM version
    if [ "${SCHEMA_VERSION}" != "${PV_SCHEMA_VERSION}" ]; then
       echo "ERROR: Something seems wrong, db schema version mismatch for the same app version: ${PV_SCHEMA_VERSION} <-> ${SCHEMA_VERSION}"
       exit 1
    fi
    # Assuming redeployment (same APP_VERSION)
    DEPLOYMENT_STATUS=redeployment
  else
    # Handle special master case
    # Master version remains static, check DB schema status and proceed accordingly

    if [[ ${APP_VERSION} == master ]]; then
       # Go for redeployment case unless rake task returns 1 (pending migrations)
       DEPLOYMENT_STATUS=redeployment
       cd ${APP_ROOT} && RAILS_USE_MEMORY_STORE=true bin/rake db:abort_if_pending_migrations
       [ "$?" -eq "1" ] && DEPLOYMENT_STATUS=upgrade
    else
     # Assuming regular upgrade (different APP_VERSION)
     # Ensure APP_VERSION must be greater than stored PV_APP_VERSION on upgrades
        check_version_gt ${APP_VERSION} ${PV_APP_VERSION}
        DEPLOYMENT_STATUS=upgrade
    fi
  fi
else
  echo "No pre-existing EVM configuration found on PV"
  DEPLOYMENT_STATUS=new_deployment
fi

}

function check_db_status() {
# Check if database pod is accepting connections

echo "== Checking DB status =="

PG_ISREADY="$(which pg_isready)"

[[ ! -x ${PG_ISREADY} ]] && echo "ERROR: Could not find pg_isready executable, aborting.." && exit 1

while true; do
  ${PG_ISREADY} -h ${DATABASE_SERVICE_NAME} && break
  sleep 5
done

}

function check_version_gt() { 
# Check if upgrade version is actually greater than stored PV version
# -V sorts alphanumeric versions within text, will always return oldest version first
# Compare sort version result against upgrade version

if [[ "$(echo "$@" | tr " " "\n" | sort -V | head -n 1)" != "$1" ]]; then
   # Version is newer return 0 and continue
   return 0

   else
     echo "ERROR: Upgrade version $1 is older than PV version $2, aborting upgrade.."
     exit 1
fi

}

function write_deployment_info() {
# Description
# Populate info file based on initial deployment and store on PV
# Output in bash format to be easily sourced
# IMAGE_VERSION is supplied by docker environment

DEPLOYMENT_DATE="$(date +%F_%T)"
APP_VERSION="$(cat ${APP_ROOT}/VERSION)"
SCHEMA_VERSION="$(cd ${APP_ROOT} && RAILS_USE_MEMORY_STORE=true bin/rake db:version | awk '{ print $3 }')"

if [[ -z $APP_VERSION || -z $SCHEMA_VERSION || -z $IMAGE_VERSION ]]; then
  echo "${PV_DEPLOY_INFO_FILE} is incomplete, one or more required variables are undefined"
  exit 1
else
  case "${DEPLOYMENT_STATUS}" in
    redeployment)
    ;;
    upgrade)
    # PV_DEPLOY_INFO_FILE must exist on upgrades
    [ ! -f "${PV_DEPLOY_INFO_FILE}" ] && echo "ERROR: Something seems wrong, ${PV_DEPLOY_INFO_FILE} could not be found" && exit 1
    # Backup existing PV_DEPLOY_INFO_FILE
    cp "${PV_DEPLOY_INFO_FILE}" "${PV_BACKUP_DIR}/backup_${PV_BACKUP_TIMESTAMP}"
    cp "${PV_DEPLOY_INFO_FILE}" "${PV_DEPLOY_INFO_FILE}~"
    # Re-write file with upgraded deployment info
    echo "PV_APP_VERSION=${APP_VERSION}" > "${PV_DEPLOY_INFO_FILE}"
    echo "PV_SCHEMA_VERSION=${SCHEMA_VERSION}" >> "${PV_DEPLOY_INFO_FILE}"
    echo "PV_IMG_VERSION=${IMAGE_VERSION}" >> "${PV_DEPLOY_INFO_FILE}"
    echo "PV_DEPLOYMENT_DATE=${DEPLOYMENT_DATE}" >> "${PV_DEPLOY_INFO_FILE}"
    ;;
    new_deployment)
    # No PV DEPLOY INFO file should exist on new deployments
    [ -f "${PV_DEPLOY_INFO_FILE}" ] && echo "ERROR: Something seems wrong, ${PV_DEPLOY_INFO_FILE} already exists on a new deployment" && exit 1
    echo "PV_APP_VERSION=${APP_VERSION}" > "${PV_DEPLOY_INFO_FILE}"
    echo "PV_SCHEMA_VERSION=${SCHEMA_VERSION}" >> "${PV_DEPLOY_INFO_FILE}"
    echo "PV_IMG_VERSION=${IMAGE_VERSION}" >> "${PV_DEPLOY_INFO_FILE}"
    echo "PV_DEPLOYMENT_DATE=${DEPLOYMENT_DATE}" >> "${PV_DEPLOY_INFO_FILE}"
    ;;
    *)
    echo "Could not find a suitable deployment status type, exiting.."
    exit 1
  esac
fi

}

function prepare_init_env() {

# Description
# Prepare appliance initialization environment

# Make a copy of CONTAINER_DATA_PERSIST_FILE into PV if not present
[ ! -f "${PV_DATA_PERSIST_FILE}" ] && cp -a "${CONTAINER_DATA_PERSIST_FILE}" "${APP_ROOT_PERSISTENT}"

# Create container deployment dirs into PV if not already present
[ ! -d "${PV_LOG_DIR}" ] && mkdir -p "${PV_LOG_DIR}"
[ ! -d "${PV_BACKUP_DIR}" ] && mkdir -p "${PV_BACKUP_DIR}"

}

function setup_logs() {
# Description
# Configure EVM logdir on PV

# Ensure EVM logdir is setup on PV before init
if [ ! -h "${APP_ROOT}/log" ]; then
  [ ! -d "${PV_CONTAINER_DATA_DIR}${APP_ROOT}/log" ] && mkdir -p "${PV_CONTAINER_DATA_DIR}${APP_ROOT}/log"
  cp -a "${APP_ROOT}/log" "${PV_CONTAINER_DATA_DIR}${APP_ROOT}"
  mv "${APP_ROOT}/log" "${APP_ROOT}/log~"
  ln --backup -sn "${PV_CONTAINER_DATA_DIR}${APP_ROOT}/log" "${APP_ROOT}/log"
fi

}

function init_appliance() {
# Description
# Execute appliance_console to initialize appliance

echo "== Initializing Appliance =="
appliance_console_cli --region ${DATABASE_REGION} --hostname ${DATABASE_SERVICE_NAME} --username ${POSTGRESQL_USER} --password ${POSTGRESQL_PASSWORD} --key

[ "$?" -ne "0" ] && echo "ERROR: Failed to initialize appliance, please check journal or appliance_console logs at ${APP_ROOT}/log/appliance_console.log" && exit 1

}

function setup_memcached() {
# Description
# Replace memcached host in EVM configuration to use assigned service pod IP

echo "== Applying memcached config =="

sed -i~ -E "s/:memcache_server:.*/:memcache_server: ${MEMCACHED_SERVICE_NAME}:11211/gi" "${APP_ROOT}/config/settings.yml"

[ "$?" -ne "0" ] && echo "ERROR: Failed to apply memcached configuration, please check journal or PV logs" && exit 1

}

function run_hook() {
# Description
# Run hook script to enable future code to be run anywhere needed (i.e upgrades)

echo "== Calling Deployment Hook =="

[[ -z $1 ]] && echo "Called hook script but no filename was provided, exiting.." && return 1

local SCRIPT_NAME="$1"

# Fixed script path and log location
HOOK_SCRIPT="${CONTAINER_SCRIPTS_ROOT}/$SCRIPT_NAME"
PV_HOOK_SCRIPT_LOG="${PV_LOG_DIR}/${SCRIPT_NAME}_hook_${PV_LOG_TIMESTAMP}.log"

(
if [ -f "${HOOK_SCRIPT}" ]; then
  # Ensure is executable
  [ ! -x "${HOOK_SCRIPT}" ] && chmod +x "${HOOK_SCRIPT}"
  # APP_VERSION and PV_APP_VERSION are set by check_deployment_status and passed to hook environment as FROM/TO vars
  echo "== Running ${HOOK_SCRIPT} =="
  FROM_VERSION="${PV_APP_VERSION}" TO_VERSION="${APP_VERSION}" "${HOOK_SCRIPT}"
  [ "$?" -ne "0" ] && echo "ERROR: ${HOOK_SCRIPT} failed, please check logs at ${PV_HOOK_SCRIPT_LOG}" && exit 1
else
  echo "Hook script ${SCRIPT_NAME} not found, skipping"
fi
) 2>&1 | tee "${PV_HOOK_SCRIPT_LOG}"

}

function migrate_db() {
# Description
# Execute DB migration, log output and check errors

PV_MIGRATE_DB_LOG="${PV_LOG_DIR}/migrate_db_${PV_LOG_TIMESTAMP}.log"

(
echo "== Migrating Database =="

cd ${APP_ROOT} && bin/rake db:migrate

[ "$?" -ne "0" ] && echo "ERROR: Failed to migrate database, please check logs at ${PV_MIGRATE_DB_LOG}" && exit 1
) 2>&1 | tee ${PV_MIGRATE_DB_LOG}

}

function sync_pv_data() {
# Description
# Process PV_DATA_PERSIST_FILE which contains the desired files/dirs to store on PV
# Use rsync to transfer files/dirs, log output and check return status

PV_DATA_SYNC_LOG="${PV_LOG_DIR}/sync_pv_data_${PV_LOG_TIMESTAMP}.log"

(
echo "== Initializing PV data =="

rsync -qavL --files-from="${PV_DATA_PERSIST_FILE}" / "${PV_CONTAINER_DATA_DIR}"

# Catch non-zero return value and print warning

[ "$?" -ne "0" ] && echo "WARNING: Some files might not have been copied please check logs at ${PV_DATA_SYNC_LOG}"

# Make database.yml, DB keys and region file are available on shared PV, rsync will create directory structure

[ ! -f "${PV_REGION_VMDB}/config/database.yml" ] && rsync -qavR "${APP_ROOT}/config/database.yml" "${PV_CONTAINER_DATA_REGION_DIR}"
[ ! -f "${PV_REGION_VMDB}/certs/v2_key" ] && rsync -qavR "${APP_ROOT}/certs/v2_key" "${PV_CONTAINER_DATA_REGION_DIR}"
[ ! -f "${PV_REGION_VMDB}/REGION" ] && rsync -qavR "${APP_ROOT}/REGION" "${PV_CONTAINER_DATA_REGION_DIR}"


) 2>&1 | tee "${PV_DATA_SYNC_LOG}"

}

function restore_pv_data() {
# Description
# Process PV_DATA_PERSIST_FILE which contains the desired files/dirs to restore from PV
# Check if file/dir exists on PV, redeploy symlinks on ${APP_ROOT} pointing to PV

PV_DATA_RESTORE_LOG="${PV_LOG_DIR}/restore_pv_data_${PV_LOG_TIMESTAMP}.log"

(
echo "== Restoring PV data symlinks =="

# Ensure PV_DATA_PERSIST_FILE is present, it should be if prepare_init_env_data was executed

[ ! -f "${PV_DATA_PERSIST_FILE}" ] && echo "ERROR: Something seems wrong, ${PV_DATA_PERSIST_FILE} was not found" && exit 1

# Ensure we always restore DB config and keys from region PV before processing PV_DATA_PERSIST_FILE, sync_pv_data populates these files

ln --backup -sn "${PV_REGION_VMDB}/config/database.yml" "${APP_ROOT}/config/database.yml"
ln --backup -sn "${PV_REGION_VMDB}/certs/v2_key" "${APP_ROOT}/certs/v2_key"
ln --backup -sn "${PV_REGION_VMDB}/REGION" "${APP_ROOT}/REGION"

while read -r FILE
do
    # Sanity checks
    [[ ${FILE} = \#* ]] && continue
    [[ ${FILE} == / ]] && continue
    [[ ! -e ${PV_CONTAINER_DATA_DIR}$FILE ]] && echo "${FILE} does not exist on PV, skipping" && continue
    [[ -h ${FILE} ]] && echo "${FILE} symlink is already in place, skipping" && continue
    # Obtain dirname and filename from source file
    DIR="$(dirname ${FILE})"
    FILENAME="$(basename ${FILE})"
    # Check if we are working with a directory, backup
    [[ -d ${FILE} ]] && mv "${FILE}" "${FILE}~"
    # Place symlink back to persistent volume
    ln --backup -sn "${PV_CONTAINER_DATA_DIR}${DIR}/${FILENAME}" "${FILE}"
done < "${PV_DATA_PERSIST_FILE}"

) 2>&1 | tee "${PV_DATA_RESTORE_LOG}"

}

function backup_pv_data() {
# Description
# Backup existing PV data before initiating an upgrade procedure

PV_DATA_BACKUP_LOG="${PV_LOG_DIR}/backup_pv_data_${PV_LOG_TIMESTAMP}.log"
PV_BACKUP_TIMESTAMP="$(date +%Y_%m_%d_%H%M%S)"

(
echo "== Initializing PV data backup =="

rsync -qav "${PV_CONTAINER_DATA_DIR}" "${PV_BACKUP_DIR}/backup_${PV_BACKUP_TIMESTAMP}"

[ "$?" -ne "0" ] && echo "WARNING: Some files might not have been copied please check logs at ${PV_DATA_BACKUP_LOG}"

) 2>&1 | tee "${PV_DATA_BACKUP_LOG}"

}
