#!/bin/bash

pushd images
  find . -name build.log | xargs rm

  docker build manageiq-base -t carbonin/manageiq-base:latest --no-cache > manageiq-base/build.log

  docker build manageiq-orchestrator -t carbonin/manageiq-orchestrator:latest > manageiq-orchestrator/build.log && docker push carbonin/manageiq-orchestrator:latest >> manageiq-orchestrator/build.log &

  docker build manageiq-base-worker -t carbonin/manageiq-base-worker:latest > manageiq-base-worker/build.log && docker push carbonin/manageiq-base-worker:latest >> manageiq-base-worker/build.log

  docker build manageiq-webserver-worker -t carbonin/manageiq-webserver-worker:latest > manageiq-webserver-worker/build.log && docker push carbonin/manageiq-webserver-worker:latest >> manageiq-webserver-worker/build.log

  docker build manageiq-ui-worker -t carbonin/manageiq-ui-worker:latest > manageiq-ui-worker/build.log && docker push carbonin/manageiq-ui-worker:latest >> manageiq-ui-worker/build.log
popd
