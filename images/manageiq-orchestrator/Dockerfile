FROM manageiq/manageiq-base:latest
MAINTAINER ManageIQ https://manageiq.org

LABEL name="manageiq-orchestrator" \
      summary="ManageIQ Orchestrator Image"

COPY container-assets/appliance-initialize.sh /usr/local/bin
COPY container-assets/check-dependent-services.sh /usr/local/bin
COPY container-assets/entrypoint /usr/local/bin

CMD ["entrypoint"]
