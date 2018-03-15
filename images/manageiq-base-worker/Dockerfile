FROM manageiq/manageiq-base:latest
MAINTAINER ManageIQ https://manageiq.org

LABEL name="manageiq-base-worker" \
      summary="ManageIQ Worker Image"

COPY container-assets/entrypoint /usr/local/bin
COPY container-assets/manageiq_liveness_check /usr/local/bin

CMD ["entrypoint"]
