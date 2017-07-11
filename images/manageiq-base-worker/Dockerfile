FROM carbonin/manageiq-base:latest
MAINTAINER ManageIQ https://manageiq.org

LABEL name="manageiq-base-worker" \
      summary="ManageIQ Worker Image"

COPY docker-assets/entrypoint /usr/local/bin
COPY docker-assets/manageiq_liveness_check /usr/local/bin

CMD ["entrypoint"]
