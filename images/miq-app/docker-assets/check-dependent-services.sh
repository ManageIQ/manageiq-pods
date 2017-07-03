#!/bin/sh

# Source OpenShift scripting env
[[ -s ${CONTAINER_SCRIPTS_ROOT}/container-deploy-common.sh ]] && source "${CONTAINER_SCRIPTS_ROOT}/container-deploy-common.sh"

# Check readiness of external services
check_svc_status ${FRONTEND_SERVICE_NAME} 80
