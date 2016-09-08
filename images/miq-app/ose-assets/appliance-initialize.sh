#!/bin/sh

[[ -s /etc/default/evm ]] && source /etc/default/evm

# Environment supplied by OpenShift MIQ via template
# Assemble DB service host variable based on template parameter $DATABASE_SERVICE_NAME

DATABASE_SVC_HOST="$(echo $DATABASE_SERVICE_NAME | tr '[:lower:]' '[:upper:]')_SERVICE_HOST"
MEMCACHED_SVC_HOST="$(echo $MEMCACHED_SERVICE_NAME | tr '[:lower:]' '[:upper:]')_SERVICE_HOST"

# Replace memcache host in EVM configuration
sed -i.bak -E "s/:memcache_server:.*/:memcache_server: ${!MEMCACHED_SVC_HOST}:11211/gi" ${APP_ROOT}/config/settings.yml

echo "Initializing Appliance, please wait ..."
appliance_console_cli --region ${DATABASE_REGION} --hostname ${!DATABASE_SVC_HOST} --username ${POSTGRESQL_USER} --password ${POSTGRESQL_PASSWORD}
