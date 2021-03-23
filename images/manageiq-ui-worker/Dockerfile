ARG FROM_REPO=manageiq
ARG FROM_TAG=latest-lasker

FROM ${FROM_REPO}/manageiq-webserver-worker:${FROM_TAG}
MAINTAINER ManageIQ https://manageiq.org

LABEL name="manageiq-ui-worker" \
      summary="ManageIQ user interface worker image"

RUN rm -f /etc/httpd/conf.d/*

# Configure httpd to run without root privileges
RUN chgrp root /var/run/httpd && chmod g+rwx /var/run/httpd && \
    chgrp root /var/log/httpd && chmod g+rwx /var/log/httpd
RUN sed -i '/^Listen 80/d' /etc/httpd/conf/httpd.conf

COPY container-assets/manageiq-http.conf /etc/httpd/conf.d

RUN source /etc/default/evm && \
    /usr/bin/generate_rpm_manifest.sh && \
    dnf clean all && \
    rm -rf /var/cache/dnf

RUN rm -rf /tmp/rpms /create_local_yum_repo.sh /etc/yum.repos.d/local_rpm.repo
