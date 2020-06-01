ARG FROM_REPO=manageiq
ARG FROM_TAG=latest

FROM ${FROM_REPO}/manageiq-base-worker:${FROM_TAG}
MAINTAINER ManageIQ https://manageiq.org

ARG RPM_PREFIX=manageiq

LABEL name="manageiq-webserver-worker" \
      summary="ManageIQ web server worker image"

RUN dnf -y --disableplugin=subscription-manager --setopt=tsflags=nodocs install ${RPM_PREFIX}-ui && dnf clean all

EXPOSE 3000
