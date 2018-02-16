FROM manageiq/manageiq-base:latest
MAINTAINER ManageIQ https://manageiq.org

LABEL name="manageiq-orchestrator" \
      summary="ManageIQ Orchestrator Image"

COPY docker-assets/appliance-initialize.sh /usr/local/bin
COPY docker-assets/check-dependent-services.sh /usr/local/bin
COPY docker-assets/entrypoint /usr/local/bin

CMD ["entrypoint"]
