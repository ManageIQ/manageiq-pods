ARG CONTAINER_PREFIX=manageiq
ARG FROM_REPO=manageiq
ARG FROM_TAG=latest

FROM ${FROM_REPO}/${CONTAINER_PREFIX}-base:${FROM_TAG}

COPY rpms/* /tmp/rpms/

RUN /create_local_yum_repo.sh && \
    dnf -y --repo=local-rpm update && \
    chgrp -R 0 $APP_ROOT && \
    chmod -R g=u $APP_ROOT && \
    source /etc/default/evm && \
    /usr/bin/generate_rpm_manifest.sh && \
    clean_dnf_rpm
