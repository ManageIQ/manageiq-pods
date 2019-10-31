#!/bin/bash

[[ -s /etc/default/evm ]] && source /etc/default/evm

function write_guid() {
  echo "${GUID}" > ${APP_ROOT}/GUID
}

function write_v2_key() {
  echo "== Writing encryption key =="
  cat > /var/www/miq/vmdb/certs/v2_key << KEY
---
:algorithm: aes-256-cbc
:key: ${V2_KEY}
KEY
}

# Check service status, requires two arguments: SVC name and SVC port (injected via template)
function check_svc_status() {
  local SVC_NAME=$1 SVC_PORT=$2

  [[ $# -lt 2 ]] && echo "Error something seems wrong, we need at least two parameters to check service status" && exit 1

  echo "== Checking ${SVC_NAME}:$SVC_PORT status =="

  while true; do
    ncat ${SVC_NAME} ${SVC_PORT} < /dev/null && break
    sleep 5
  done
  echo "${SVC_NAME}:${SVC_PORT} - accepting connections"
}
