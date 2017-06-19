#!/bin/sh

[[ -s /etc/default/evm ]] && source /etc/default/evm

# Source OpenShift scripting env
[[ -s ${CONTAINER_SCRIPTS_ROOT}/container-deploy-common.sh ]] && source "${CONTAINER_SCRIPTS_ROOT}/container-deploy-common.sh"

# Delay in seconds before we init, allows rest of services to settle
sleep "${APPLICATION_INIT_DELAY}"

# Prepare initialization environment
prepare_init_env

# Check Memcached readiness
check_svc_status ${MEMCACHED_SERVICE_NAME} 11211

setup_memcached

# Check DB readiness
check_svc_status ${DATABASE_SERVICE_NAME} 5432

write_v2_key

cd ${APP_ROOT}
bin/rake evm:deployment_status

# Select path of action based on DEPLOYMENT_STATUS value
case $? in
  0) # new_deployment
    echo "== Starting New Deployment =="
    # Generate the certs
    /usr/bin/generate_miq_server_cert.sh

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
  1) # new_replica
    echo "== Starting New Replica =="
    /usr/bin/generate_miq_server_cert.sh
    setup_logs
    replica_join_region
    sync_pv_data
    restore_pv_data
  ;;
  2) # redeployment
    echo "== Starting Re-deployment =="
    restore_pv_data
  ;;
  3) # upgrade
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
