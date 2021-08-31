ARG FROM_REPO=manageiq
ARG FROM_TAG=latest-morphy-lasker

FROM ${FROM_REPO}/manageiq-base-worker:${FROM_TAG}
MAINTAINER ManageIQ https://manageiq.org

ARG RPM_PREFIX=manageiq

LABEL name="manageiq-webserver-worker" \
      summary="ManageIQ web server worker image"

RUN dnf -y --disableplugin=subscription-manager --setopt=tsflags=nodocs install ${RPM_PREFIX}-ui && \
    dnf clean all && \
    rm -rf /var/cache/dnf

RUN source /etc/default/evm && \
    /usr/bin/generate_rpm_manifest.sh && \
    dnf clean all && \
    rm -rf /var/cache/dnf

EXPOSE 3000
