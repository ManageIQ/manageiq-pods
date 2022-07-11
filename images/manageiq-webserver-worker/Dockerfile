ARG FROM_REPO=manageiq
ARG FROM_TAG=latest

FROM ${FROM_REPO}/manageiq-base-worker:${FROM_TAG}
MAINTAINER ManageIQ https://manageiq.org

ARG RPM_PREFIX=manageiq

LABEL name="manageiq-webserver-worker" \
      summary="ManageIQ web server worker image"

COPY container-assets/service-worker-entrypoint /usr/local/bin

RUN dnf -y install \
      ${RPM_PREFIX}-ui && \
    clean_dnf_rpm && \
    # Remove httpd default settings
    rm -f /etc/httpd/conf.d/* && \
    sed -i '/^Listen 80/d' /etc/httpd/conf/httpd.conf && \
    # Configure httpd to run without root privileges
    chgrp root /var/run/httpd && \
    chmod g+rwx /var/run/httpd && \
    chgrp root /var/log/httpd && \
    chmod g+rwx /var/log/httpd

# Build the RPM manifest
RUN source /etc/default/evm && \
    /usr/bin/generate_rpm_manifest.sh && \
    clean_dnf_rpm

EXPOSE 3000
EXPOSE 4000

CMD ["service-worker-entrypoint"]
