#!/bin/bash
# This script is intended to be run in appliance-initialize systemd unit stop-post

[[ -s /etc/default/evm ]] && source /etc/default/evm

[[ -s ${CONTAINER_SCRIPTS_ROOT}/container-deploy-common.sh ]] && source ${CONTAINER_SCRIPTS_ROOT}/container-deploy-common.sh

# Dump journal logs into PV (ensure log access for troubleshooting purposes)

/usr/bin/journalctl -u appliance-initialize > ${PV_LOG_DIR}/appliance_initialize_${PV_LOG_TIMESTAMP}.log
