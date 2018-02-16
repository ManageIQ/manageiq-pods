#!/bin/bash

pushd images
  find . -name build.log | xargs rm

  docker build manageiq-base -t manageiq/manageiq-base:latest --no-cache > manageiq-base/build.log

  docker build manageiq-orchestrator -t manageiq/manageiq-orchestrator:latest > manageiq-orchestrator/build.log && docker push manageiq/manageiq-orchestrator:latest >> manageiq-orchestrator/build.log &

  docker build manageiq-base-worker -t manageiq/manageiq-base-worker:latest > manageiq-base-worker/build.log && docker push manageiq/manageiq-base-worker:latest >> manageiq-base-worker/build.log

  docker build manageiq-webserver-worker -t manageiq/manageiq-webserver-worker:latest > manageiq-webserver-worker/build.log && docker push manageiq/manageiq-webserver-worker:latest >> manageiq-webserver-worker/build.log

  docker build manageiq-ui-worker -t manageiq/manageiq-ui-worker:latest > manageiq-ui-worker/build.log && docker push manageiq/manageiq-ui-worker:latest >> manageiq-ui-worker/build.log
popd
