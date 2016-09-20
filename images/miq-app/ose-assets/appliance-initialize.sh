#!/bin/sh

[[ -s /etc/default/evm ]] && source /etc/default/evm

# Ensure OpenShift scripting environment is present, exit if source fails
[[ -s ${SCRIPTS_ROOT}/ose-deploy-common.sh ]] && source ${SCRIPTS_ROOT}/ose-deploy-common.sh || { echo "Failed to source ${SCRIPTS_ROOT}/ose-deploy-common.sh" ; exit 1; }

# Prepare initialization environment
prepare_init_env

# Check deployment status
check_deployment_status

if [ -n "${UPGRADE}" ]; then
  echo "== Starting Upgrade =="
  restore_pv_data
  setup_memcached
  migrate_db
fi

if [ -n "${REDEPLOY}" ]; then
  echo "== Starting Re-deployment =="
  restore_pv_data
  setup_memcached
  migrate_db
fi

if [ -n "${NEW_DEPLOYMENT}" ]; then

  # Setup logs on PV
  setup_logs

  # Setup memcached host in EVM configuration
  setup_memcached
 
  # Initialize EVM appliance 
  echo "== Initializing Appliance =="
  appliance_console_cli --region ${DATABASE_REGION} --hostname ${DATABASE_SERVICE_NAME} --username ${POSTGRESQL_USER} --password ${POSTGRESQL_PASSWORD}

  # Sync persistent data on PV
  sync_pv_data

  # Restore symlinks from PV data
  restore_pv_data

  # Write deployment info file to PV
  write_deployment_info
fi
