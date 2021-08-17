ARG FROM_REPO=manageiq
ARG FROM_TAG=latest

FROM ${FROM_REPO}/manageiq-webserver-worker:${FROM_TAG}
MAINTAINER ManageIQ https://manageiq.org

LABEL name="manageiq-ui-worker" \
      summary="ManageIQ user interface worker image"

RUN source /etc/default/evm && \
    /usr/bin/generate_rpm_manifest.sh && \
    dnf clean all && \
    rm -rf /var/cache/dnf

RUN rm -rf /tmp/rpms /create_local_yum_repo.sh /etc/yum.repos.d/local_rpm.repo
