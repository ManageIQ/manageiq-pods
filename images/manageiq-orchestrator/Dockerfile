ARG FROM_REPO=docker.io/manageiq
ARG FROM_TAG=latest

FROM ${FROM_REPO}/manageiq-base:${FROM_TAG}

LABEL name="manageiq-orchestrator" \
      summary="ManageIQ Orchestrator Image"

COPY container-assets/entrypoint /usr/local/bin

CMD ["entrypoint"]
