#!/bin/bash

[[ -s /etc/default/evm ]] && source /etc/default/evm

# This directory is used to store server specific data to be persisted
PV_CONTAINER_DATA_DIR="${APP_ROOT_PERSISTENT}/server-data"

# This directory is used to store server specific container deployment data (logs,backups,etc)
PV_CONTAINER_DEPLOY_DIR="${APP_ROOT_PERSISTENT}/server-deploy"

# This directory is used to store server specific initialization logfiles on PV
PV_LOG_DIR="${PV_CONTAINER_DEPLOY_DIR}/log"

# Directory used to backup server specific PV data before performing an upgrade
PV_BACKUP_DIR="${PV_CONTAINER_DEPLOY_DIR}/backup"

# This file is supplied by the app container image with default files/dirs to persist on PV
DATA_PERSIST_FILE="/container.data.persist"

# Set log timestamp for running instance
PV_LOG_TIMESTAMP="$(date +%s)"

function write_v2_key() {
  echo "== Writing encryption key =="
  cat > /var/www/miq/vmdb/certs/v2_key << KEY
---
:algorithm: aes-256-cbc
:key: ${V2_KEY}
KEY
}

# Check service status, requires two arguments: SVC name and SVC port (injected via template)
function check_svc_status() {
  NCAT="$(which ncat)"
  local SVC_NAME=$1 SVC_PORT=$2

  [[ $# -lt 2 ]] && echo "Error something seems wrong, we need at least two parameters to check service status" && exit 1

  echo "== Checking ${SVC_NAME}:$SVC_PORT status =="

  [[ ! -x ${NCAT} ]] && echo "ERROR: Could not find ncat executable, aborting.." && exit 1

  while true; do
    ${NCAT} ${SVC_NAME} ${SVC_PORT} < /dev/null && break
    sleep 5
  done
  echo "${SVC_NAME}:${SVC_PORT} - accepting connections"
}

# Join the new server/replica to the remote region
function replica_join_region() {
  echo "== Joining region =="
  cd ${APP_ROOT} && RAILS_USE_MEMORY_STORE=true bin/rake evm:join_region
}

# Prepare appliance initialization environment
function prepare_init_env() {
  # Create container deployment dirs into PV if not already present
  [ ! -d "${PV_LOG_DIR}" ] && mkdir -p "${PV_LOG_DIR}"
  [ ! -d "${PV_BACKUP_DIR}" ] && mkdir -p "${PV_BACKUP_DIR}"
}

# Configure EVM logdir on PV
function setup_logs() {
  # Ensure EVM logdir is setup on PV before init
  if [ ! -h "${APP_ROOT}/log" ]; then
    [ ! -d "${PV_CONTAINER_DATA_DIR}${APP_ROOT}/log" ] && mkdir -p "${PV_CONTAINER_DATA_DIR}${APP_ROOT}/log"
    cp -a "${APP_ROOT}/log" "${PV_CONTAINER_DATA_DIR}${APP_ROOT}"
    mv "${APP_ROOT}/log" "${APP_ROOT}/log~"
    ln --backup -sn "${PV_CONTAINER_DATA_DIR}${APP_ROOT}/log" "${APP_ROOT}/log"
  fi
}

# Execute appliance_console to initialize appliance
function init_appliance() {
  echo "== Initializing Appliance =="

  pushd ${APP_ROOT}
    bundle exec rake evm:db:region -- --region ${DATABASE_REGION}
  popd

  [ "$?" -ne "0" ] && echo "ERROR: Failed to initialize appliance, please check journal or appliance_console logs at ${APP_ROOT}/log/appliance_console.log" && exit 1
}

# Run hook script to enable future code to be run anywhere needed (i.e upgrades)
function run_hook() {
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
      echo "== Running ${HOOK_SCRIPT} =="
      ${HOOK_SCRIPT}
      [ "$?" -ne "0" ] && echo "ERROR: ${HOOK_SCRIPT} failed, please check logs at ${PV_HOOK_SCRIPT_LOG}" && exit 1
    else
      echo "Hook script ${SCRIPT_NAME} not found, skipping"
    fi
  ) 2>&1 | tee "${PV_HOOK_SCRIPT_LOG}"
}

# Execute DB migration, log output and check errors
function migrate_db() {
  PV_MIGRATE_DB_LOG="${PV_LOG_DIR}/migrate_db_${PV_LOG_TIMESTAMP}.log"

  (
    echo "== Migrating Database =="

    cd ${APP_ROOT} && bin/rake db:migrate

    [ "$?" -ne "0" ] && echo "ERROR: Failed to migrate database, please check logs at ${PV_MIGRATE_DB_LOG}" && exit 1
  ) 2>&1 | tee ${PV_MIGRATE_DB_LOG}
}

# Set EVM admin pwd
function set_admin_pwd() {
 echo "== Setting admin password =="

   cd ${APP_ROOT} && bin/rails runner "EvmDatabase.seed_primordial; user = User.find_by_userid('admin').update_attributes!(:password => ENV['APPLICATION_ADMIN_PASSWORD'])"

   [ "$?" -ne "0" ] && echo "ERROR: Failed to set admin password, please check appliance logs"
}

# Process DATA_PERSIST_FILE which contains the desired files/dirs to store on the PV
# Use rsync to transfer files/dirs, log output and check return status
# Ensure we always store an initial data backup on PV
function init_pv_data() {
  PV_DATA_INIT_LOG="${PV_LOG_DIR}/init_pv_data_${PV_LOG_TIMESTAMP}.log"

  (
    echo "== Initializing PV data =="

    sync_pv_data

  ) 2>&1 | tee "${PV_DATA_INIT_LOG}"
}

# Process DATA_PERSIST_FILE which contains the desired files/dirs to restore from PV
# Check if file/dir exists on PV, redeploy symlinks on ${APP_ROOT} pointing to PV
function restore_pv_data() {
  PV_DATA_RESTORE_LOG="${PV_LOG_DIR}/restore_pv_data_${PV_LOG_TIMESTAMP}.log"

  (
    echo "== Restoring PV data symlinks =="
    while read -r FILE
    do
      # Sanity checks
      [[ ${FILE} = \#* ]] && continue
      [[ ${FILE} == / ]] && continue
      [[ -h ${FILE} ]] && echo "${FILE} symlink is already in place, skipping" && continue
      [[ ! -e ${PV_CONTAINER_DATA_DIR}$FILE ]] && echo "${FILE} does not exist on PV, skipping" && continue
      # Obtain dirname and filename from source file
      DIR="$(dirname ${FILE})"
      FILENAME="$(basename ${FILE})"
      # Check if we are working with a directory, backup
      [[ -d ${FILE} ]] && mv "${FILE}" "${FILE}~"
      # Place symlink back to persistent volume
      ln --backup -sn "${PV_CONTAINER_DATA_DIR}${DIR}/${FILENAME}" "${FILE}"
    done < "${DATA_PERSIST_FILE}"
  ) 2>&1 | tee "${PV_DATA_RESTORE_LOG}"
}

# Backup existing PV data before initiating an upgrade procedure
# Exclude EVM server logs
function backup_pv_data() {
  PV_DATA_BACKUP_LOG="${PV_LOG_DIR}/backup_pv_data_${PV_LOG_TIMESTAMP}.log"
  PV_BACKUP_TIMESTAMP="$(date +%Y_%m_%d_%H%M%S)"

  (
    echo "== Initializing PV data backup =="

    rsync -av --exclude 'log' "${PV_CONTAINER_DATA_DIR}" "${PV_BACKUP_DIR}/backup_${PV_BACKUP_TIMESTAMP}"

    [ "$?" -ne "0" ] && echo "WARNING: Some files might not have been copied please check logs at ${PV_DATA_BACKUP_LOG}"
  ) 2>&1 | tee "${PV_DATA_BACKUP_LOG}"
}

# Process data persist file and sync data back to PV
# Skip essential files that should never be synced back after region initialization
function sync_pv_data() {
  PV_DATA_SYNC_LOG="${PV_LOG_DIR}/sync_pv_data_${PV_LOG_TIMESTAMP}.log"

  (
    echo "== Syncing persist data to server PV =="

    rsync -qavL --files-from="${DATA_PERSIST_FILE}" / "${PV_CONTAINER_DATA_DIR}"

    [ "$?" -ne "0" ] && echo "WARNING: Some files might not have been copied please check logs at ${PV_DATA_SYNC_LOG}"
  ) 2>&1 | tee "${PV_DATA_SYNC_LOG}"
}
