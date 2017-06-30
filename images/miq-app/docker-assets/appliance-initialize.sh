#!/bin/sh

[[ -s /etc/default/evm ]] && source /etc/default/evm

# Source OpenShift scripting env
[[ -s ${CONTAINER_SCRIPTS_ROOT}/container-deploy-common.sh ]] && source "${CONTAINER_SCRIPTS_ROOT}/container-deploy-common.sh"

# Delay in seconds before we init, allows rest of services to settle
sleep "${APPLICATION_INIT_DELAY}"

# Prepare initialization environment
prepare_init_env

write_v2_key

restore_pv_data

# Generate httpd certificate
/usr/bin/generate_miq_server_cert.sh

cd ${APP_ROOT}
bin/rake evm:deployment_status
case $? in
  3) # new_deployment
    echo "== Starting New Deployment =="
    # Setup logs on PV before init
    setup_logs

    # Run appliance_console_cli to init appliance
    init_appliance

    # Init persistent data from application rootdir on PV
    init_pv_data

    # Make initial backup
    backup_pv_data

    # Restore symlinks from PV to application rootdir
    restore_pv_data
  ;;
  4) # new_replica
    echo "== Starting New Replica =="
    setup_logs
    replica_join_region
    sync_pv_data
    restore_pv_data
  ;;
  5) # redeployment
    echo "== Starting Re-deployment =="
  ;;
  6) # upgrade
    echo "== Starting Upgrade =="
    backup_pv_data
    run_hook pre-upgrade
    restore_pv_data
    migrate_db
    run_hook post-upgrade
  ;;
  *)
    echo "Could not find a suitable deployment type, exiting.."
    exit 1
esac
