#!/bin/bash

[[ -s /etc/default/evm ]] && source /etc/default/evm

function write_guid() {
  echo "${GUID}" > ${APP_ROOT}/GUID
}

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

# Execute appliance_console to initialize appliance
function init_appliance() {
  echo "== Initializing Appliance =="

  pushd ${APP_ROOT}
    REGION=${DATABASE_REGION} bundle exec rake db:migrate
    REGION=${DATABASE_REGION} bundle exec rails runner "MiqDatabase.seed; MiqRegion.seed"
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

  if [ -f "${HOOK_SCRIPT}" ]; then
    # Ensure is executable
    [ ! -x "${HOOK_SCRIPT}" ] && chmod +x "${HOOK_SCRIPT}"
    echo "== Running ${HOOK_SCRIPT} =="
    ${HOOK_SCRIPT}
    [ "$?" -ne "0" ] && echo "ERROR: ${HOOK_SCRIPT} failed" && exit 1
  else
    echo "Hook script ${SCRIPT_NAME} not found, skipping"
  fi
}

# Set EVM admin pwd
function set_admin_pwd() {
 echo "== Setting admin password =="

   cd ${APP_ROOT} && bin/rails runner "EvmDatabase.seed_primordial; user = User.find_by_userid('admin').update_attributes!(:password => ENV['APPLICATION_ADMIN_PASSWORD'])"

   [ "$?" -ne "0" ] && echo "ERROR: Failed to set admin password, please check appliance logs"
}

# Execute DB migration, log output and check errors
function migrate_db() {
  echo "== Migrating Database =="

  cd ${APP_ROOT} && bin/rake db:migrate

  [ "$?" -ne "0" ] && echo "ERROR: Failed to migrate database" && exit 1
}
