#!/bin/bash

[[ -s /etc/default/evm ]] && source /etc/default/evm

# This script is intended to run in an OpenShift postStart lifecycle hook
# All of OpenShift service and template environment variables are dropped by systemd
# Dump required ENV on a file, appliance-initialize will source before begin initialization

env | grep -e POSTGRES -e SERVICE > ${APP_ROOT}/tmp/ose-env
env | grep -e MEMCACHE -e SERVICE >>${APP_ROOT}/tmp/ose-env
