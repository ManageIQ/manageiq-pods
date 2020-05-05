ARG FROM_REPO=manageiq
ARG FROM_TAG=latest

FROM ${FROM_REPO}/manageiq-webserver-worker:${FROM_TAG}
MAINTAINER ManageIQ https://manageiq.org

ARG ARCH=x86_64

LABEL name="manageiq-ui-worker" \
      summary="ManageIQ user interface worker image"

# Install httpd and remove the existing httpd config from manageiq-appliance
RUN dnf -y --disableplugin=subscription-manager --setopt=tsflags=nodocs install \
      httpd \
      manageiq-ui && \
    rm -f /etc/httpd/conf.d/* && \
    dnf clean all

# Configure httpd to run without root privileges
RUN chgrp root /var/run/httpd && chmod g+rwx /var/run/httpd && \
    chgrp root /var/log/httpd && chmod g+rwx /var/log/httpd
RUN sed -i '/^Listen 80/d' /etc/httpd/conf/httpd.conf

COPY container-assets/manageiq-http.conf /etc/httpd/conf.d
