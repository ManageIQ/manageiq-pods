ARG FROM_REPO=manageiq
ARG FROM_TAG=latest-kasparov

FROM ${FROM_REPO}/manageiq-base:${FROM_TAG}
MAINTAINER ManageIQ https://manageiq.org

LABEL name="manageiq-base-worker" \
      summary="ManageIQ Worker Image"

COPY container-assets/entrypoint /usr/local/bin
COPY container-assets/manageiq_liveness_check /usr/local/bin

RUN source /etc/default/evm && /usr/bin/generate_rpm_manifest.sh

CMD ["entrypoint"]
