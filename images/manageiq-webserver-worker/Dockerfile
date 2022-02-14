ARG FROM_REPO=manageiq
ARG FROM_TAG=latest-najdorf

FROM ${FROM_REPO}/manageiq-base-worker:${FROM_TAG}
MAINTAINER ManageIQ https://manageiq.org

ARG RPM_PREFIX=manageiq

LABEL name="manageiq-webserver-worker" \
      summary="ManageIQ web server worker image"

RUN dnf -y --setopt=tsflags=nodocs install ${RPM_PREFIX}-ui && \
    dnf clean all && \
    rm -rf /var/cache/dnf

COPY container-assets/service-worker-entrypoint /usr/local/bin

RUN rm -f /etc/httpd/conf.d/* && \
    sed -i '/^Listen 80/d' /etc/httpd/conf/httpd.conf

# Configure httpd to run without root privileges
RUN chgrp root /var/run/httpd && chmod g+rwx /var/run/httpd && \
    chgrp root /var/log/httpd && chmod g+rwx /var/log/httpd

RUN source /etc/default/evm && \
    /usr/bin/generate_rpm_manifest.sh && \
    dnf clean all && \
    rm -rf /var/cache/dnf

EXPOSE 3000

CMD ["service-worker-entrypoint"]
