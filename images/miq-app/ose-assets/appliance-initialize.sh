#!/bin/sh

[[ -s /etc/default/evm ]] && source /etc/default/evm

# Ensure OpenShift scripting environment is present, exit if source fails
[[ -s ${SCRIPTS_ROOT}/ose-deploy-common.sh ]] && source ${SCRIPTS_ROOT}/ose-deploy-common.sh || { echo "Failed to source ${SCRIPTS_ROOT}/ose-deploy-common.sh" ; exit 1; }

# Prepare service host vars
prepare_svc_vars

# Check deployment status
check_deployment_status

if [ -n "${UPGRADE}" ]; then
  echo "== Starting Upgrade =="
  echo "== Migrating Database =="
  cd ${APP_ROOT} && bin/rake db:migrate
fi

if [ -n "${REDEPLOY}" ]; then
  echo "== Starting Re-deployment =="
  echo "== Migrating Database =="
  cd ${APP_ROOT} && bin/rake db:migrate
fi

if [ -n "${NEW_DEPLOYMENT}" ]; then

  # Setup memcached host in EVM configuration
  setup_memcached

  # Setup persistent data files/dirs on PV
  setup_persistent_data
 
  # Initialize EVM appliance 
  echo "== Initializing Appliance =="
  appliance_console_cli --region ${DATABASE_REGION} --hostname ${!DATABASE_SVC_HOST} --username ${POSTGRESQL_USER} --password ${POSTGRESQL_PASSWORD}

  # Write deployment info file to PV
  write_deployment_info
fi
