ARG FROM_REPO=docker.io/manageiq
ARG FROM_TAG=latest

# alias stage for rpms
FROM ${FROM_REPO}/manageiq-rpms:${FROM_TAG} as rpms

################################################################################

FROM registry.access.redhat.com/ubi9/ubi
ARG FROM_REPO
ARG FROM_TAG
ARG RELEASE_BUILD
ARG RPM_PREFIX=manageiq

ENV TERM=xterm \
    CONTAINER=true \
    APP_ROOT=/var/www/miq/vmdb

LABEL name="manageiq-base" \
      summary="ManageIQ base application image" \
      vendor="ManageIQ" \
      url="https://manageiq.org/" \
      description="ManageIQ is a management and automation platform for virtual, private, and hybrid cloud infrastructures." \
      io.k8s.display-name="ManageIQ" \
      io.k8s.description="ManageIQ is a management and automation platform for virtual, private, and hybrid cloud infrastructures." \
      io.openshift.tags="ManageIQ,miq,manageiq"

# Force the sticky bit on /tmp - https://bugzilla.redhat.com/show_bug.cgi?id=2138434
RUN chmod +t /tmp

RUN chmod -R g+w /etc/pki/ca-trust && \
    chmod -R g+w /usr/share/pki/ca-trust-legacy

# Install dumb-init to be used as the entrypoint
RUN ARCH=$(uname -m) && \
    curl -L -o /usr/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_${ARCH} && \
    chmod +x /usr/bin/dumb-init

COPY container-assets/prepare_local_yum_repo.sh /usr/local/bin
COPY container-assets/clean_dnf_rpm /usr/local/bin/

RUN --mount=type=bind,from=rpms,source=/tmp/rpms,target=/tmp/rpms \
    ARCH=$(uname -m) && \
    dnf config-manager --setopt=tsflags=nodocs --setopt=install_weak_deps=False --save && \
    dnf -y --setopt=protected_packages= remove redhat-release && \
    dnf -y remove *subscription-manager* && \
    dnf -y install \
      http://mirror.stream.centos.org/9-stream/BaseOS/${ARCH}/os/Packages/centos-stream-release-9.0-24.el9.noarch.rpm \
      http://mirror.stream.centos.org/9-stream/BaseOS/${ARCH}/os/Packages/centos-stream-repos-9.0-24.el9.noarch.rpm \
      http://mirror.stream.centos.org/9-stream/BaseOS/${ARCH}/os/Packages/centos-gpg-keys-9.0-24.el9.noarch.rpm && \
    dnf config-manager --enable crb && \
    dnf -y install \
      https://dl.fedoraproject.org/pub/epel/epel-release-latest-9.noarch.rpm \
      https://rpm.manageiq.org/release/19-spassky/el9/noarch/manageiq-release-19.0-1.el9.noarch.rpm && \
    if [[ "$RELEASE_BUILD" != "true" ]]; then dnf config-manager --enable manageiq-19-spassky-nightly; fi && \
    if [[ -n "$(ls /tmp/rpms)" ]]; then /usr/local/bin/prepare_local_yum_repo.sh; fi && \
    dnf -y module enable ruby:3.1 && \
    dnf -y install \
      httpd \
      mod_ssl \
      ${RPM_PREFIX}-pods \
      python3-devel && \
    clean_dnf_rpm && \
    chown manageiq:root /home/manageiq && \
    chmod -R g=u /home/manageiq && \
    chgrp -R 0 $APP_ROOT && \
    chmod -R g=u $APP_ROOT

# Add in the container_env file now that the APP_ROOT is created from the RPM
ADD container-assets/container_env ${APP_ROOT}/

# Build the RPM manifest
RUN --mount=type=bind,from=rpms,source=/tmp/rpms,target=/tmp/rpms \
    source /etc/default/evm && \
    /usr/bin/generate_rpm_manifest.sh && \
    clean_dnf_rpm

ENTRYPOINT ["/usr/bin/dumb-init", "--single-child", "--"]
