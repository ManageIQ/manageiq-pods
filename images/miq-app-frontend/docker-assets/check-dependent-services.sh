#!/bin/sh

# Source OpenShift scripting env
[[ -s ${CONTAINER_SCRIPTS_ROOT}/container-deploy-common.sh ]] && source "${CONTAINER_SCRIPTS_ROOT}/container-deploy-common.sh"

# Check readiness of external services
check_svc_status ${MEMCACHED_SERVICE_NAME} 11211
check_svc_status ${DATABASE_SERVICE_NAME} 5432
