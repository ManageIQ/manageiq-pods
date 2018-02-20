#!/bin/sh

[[ -s /etc/default/evm ]] && source /etc/default/evm

# Source OpenShift scripting env
[[ -s ${CONTAINER_SCRIPTS_ROOT}/container-deploy-common.sh ]] && source "${CONTAINER_SCRIPTS_ROOT}/container-deploy-common.sh"

write_v2_key

write_guid

cd ${APP_ROOT}
bin/rake evm:deployment_status
case $? in
  3) # new_deployment
    echo "== Starting New Deployment =="
    # Run appliance_console_cli to init appliance
    init_appliance
    set_admin_pwd
  ;;
  4) # new_replica
    echo "New replica is not supported, exiting.."
    exit 1
  ;;
  5) # redeployment
    echo "== Starting Re-deployment =="
  ;;
  6) # upgrade
    echo "== Starting Upgrade =="
    run_hook pre-upgrade
    migrate_db
    run_hook post-upgrade
  ;;
  *)
    echo "Could not find a suitable deployment type, exiting.."
    exit 1
esac
