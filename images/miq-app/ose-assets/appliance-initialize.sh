#!/bin/sh

[[ -s /etc/default/evm ]] && source /etc/default/evm

# Ensure OpenShift scripting environment is present, exit if source fails
[[ -s ${SCRIPTS_ROOT}/ose-deploy-common.sh ]] && source ${SCRIPTS_ROOT}/ose-deploy-common.sh || { echo "Failed to source ${SCRIPTS_ROOT}/ose-deploy-common.sh" ; exit 1; }

# Check deployment status (new, upgrade or redeploy)

check_deployment_status

if [ -n "${UPGRADE}" ]; then
  echo "Deployment upgrade (do something) : ${UPGRADE}"
fi

if [ -n "${REDEPLOY}" ]; then
  echo "This is a redeployment (do something) : ${REDEPLOY}"
fi

if [ -n "${NEW_DEPLOYMENT}" ]; then

  # Environment supplied by OpenShift MIQ via template
  # Assemble DB service host variable based on template parameter $DATABASE_SERVICE_NAME

  DATABASE_SVC_HOST="$(echo $DATABASE_SERVICE_NAME | tr '[:lower:]' '[:upper:]')_SERVICE_HOST"
  MEMCACHED_SVC_HOST="$(echo $MEMCACHED_SERVICE_NAME | tr '[:lower:]' '[:upper:]')_SERVICE_HOST"

  # Replace memcache host in EVM configuration
  sed -i.bak -E "s/:memcache_server:.*/:memcache_server: ${!MEMCACHED_SVC_HOST}:11211/gi" ${APP_ROOT}/config/settings.yml

  echo "Initializing Appliance, please wait ..."
  appliance_console_cli --region ${DATABASE_REGION} --hostname ${!DATABASE_SVC_HOST} --username ${POSTGRESQL_USER} --password ${POSTGRESQL_PASSWORD}
  write_deployment_info
fi
