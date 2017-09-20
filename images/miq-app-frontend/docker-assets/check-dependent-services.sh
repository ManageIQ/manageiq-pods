#!/bin/sh

# Source OpenShift scripting env
[[ -s ${CONTAINER_SCRIPTS_ROOT}/container-deploy-common.sh ]] && source "${CONTAINER_SCRIPTS_ROOT}/container-deploy-common.sh"

# Check readiness of external services
check_svc_status ${MEMCACHED_SERVICE_HOST} ${MEMCACHED_SERVICE_PORT}
check_svc_status ${POSTGRESQL_SERVICE_HOST} ${POSTGRESQL_SERVICE_PORT}
