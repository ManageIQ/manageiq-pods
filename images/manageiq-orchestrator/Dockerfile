ARG FROM_REPO=manageiq
ARG FROM_TAG=latest

FROM ${FROM_REPO}/manageiq-base:${FROM_TAG}
MAINTAINER ManageIQ https://manageiq.org

LABEL name="manageiq-orchestrator" \
      summary="ManageIQ Orchestrator Image"

COPY container-assets/entrypoint /usr/local/bin

RUN source /etc/default/evm && \
    /usr/bin/generate_rpm_manifest.sh && \
    dnf clean all && \
    rm -rf /var/cache/dnf

CMD ["entrypoint"]
