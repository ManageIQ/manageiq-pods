#!/bin/sh

[[ -s /etc/default/evm ]] && source /etc/default/evm

# $DATABASE_SERVICE_NAME AND MEMCACHED_SERVICE_NAME are not initialized by default
#  when using the docker-compose file: contrib/docker-compose.yml
# this is to avoir problem if the env variable parameters of the miq container is removed by mistake
# Note: there may be a better way to this
if [[ -z $DATABASE_SERVICE_NAME ]]; then
  DATABASE_SERVICE_NAME='postgresql'
fi
if [[ -z $MEMCACHED_SERVICE_NAME ]]; then
  MEMCACHED_SERVICE_NAME='memcached'
fi


# $POSTGRESQL_USER and $POSTGRESQL_PASSWORD are supplied by OpenShift MIQ template
# Assemble DB service host variable based on template parameter $DATABASE_SERVICE_NAME

DB_SVC_HOST="$(echo $DATABASE_SERVICE_NAME | tr '[:lower:]' '[:upper:]')_SERVICE_HOST"
MEMCACHED_SVC_HOST="$(echo $MEMCACHED_SERVICE_NAME | tr '[:lower:]' '[:upper:]')_SERVICE_HOST"

# if {MEMCACHED}_SERVICE_HOST is empty, force service name to 'memcached'
if [[ -z ${!MEMCACHED_SVC_HOST} ]]; then
  eval ${MEMCACHED_SVC_HOST}='memcached'
fi

# if {DATABASE}_SERVICE_HOST is empty, force service name to 'postgresql'
if [[ -z ${!DB_SVC_HOST} ]]; then
  eval ${DB_SVC_HOST}='postgresql'
fi

# replace memcache host in configuration
sed -i.bak -E "s/:memcache_server:.*/:memcache_server: ${!MEMCACHED_SVC_HOST}:11211/gi" ${APP_ROOT}/config/settings.yml

echo "Initializing Appliance, please wait ..."
appliance_console_cli --region 0 --hostname ${!DB_SVC_HOST} --username ${POSTGRESQL_USER} --password ${POSTGRESQL_PASSWORD}
