ARG FROM_REPO=docker.io/manageiq
ARG FROM_TAG=latest-spassky
ARG FROM_IMAGE=${FROM_REPO}/manageiq-base:${FROM_TAG}

FROM ${FROM_IMAGE}

COPY rpms/* /tmp/rpms/

RUN dnf -y install createrepo_c && \
    rm -rf /tmp/rpms/repodata && \
    /create_local_yum_repo.sh && \
    dnf -y --repo=local-rpm update && \
    chgrp -R 0 $APP_ROOT && \
    chmod -R g=u $APP_ROOT && \
    source /etc/default/evm && \
    /usr/bin/generate_rpm_manifest.sh && \
    clean_dnf_rpm
