FROM registry.access.redhat.com/ubi8/ubi
MAINTAINER ManageIQ https://manageiq.org

ARG ARCH=x86_64
ARG LOCAL_RPM
ARG RELEASE_BUILD
ARG RPM_PREFIX=manageiq

ENV TERM=xterm \
    CONTAINER=true \
    APP_ROOT=/var/www/miq/vmdb

LABEL name="manageiq-base" \
      vendor="ManageIQ" \
      url="https://manageiq.org/" \
      summary="ManageIQ base application image" \
      description="ManageIQ is a management and automation platform for virtual, private, and hybrid cloud infrastructures." \
      io.k8s.display-name="ManageIQ" \
      io.k8s.description="ManageIQ is a management and automation platform for virtual, private, and hybrid cloud infrastructures." \
      io.openshift.tags="ManageIQ,miq,manageiq"

# Force the sticky bit on /tmp - https://bugzilla.redhat.com/show_bug.cgi?id=2138434
RUN chmod +t /tmp

RUN chmod -R g+w /etc/pki/ca-trust && \
    chmod -R g+w /usr/share/pki/ca-trust-legacy

# Install dumb-init to be used as the entrypoint
RUN curl -L -o /usr/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_${ARCH} && \
    chmod +x /usr/bin/dumb-init

COPY rpms/* /tmp/rpms/
COPY container-assets/create_local_yum_repo.sh /
COPY container-assets/clean_dnf_rpm /usr/local/bin/

RUN dnf config-manager --setopt=tsflags=nodocs --setopt=install_weak_deps=False --save && \
    if [ ${ARCH} != "s390x" ] ; then \
      dnf -y --setopt=protected_packages= remove redhat-release && \
      dnf -y remove *subscription-manager* && \
      dnf -y install \
        http://mirror.centos.org/centos/8-stream/BaseOS/${ARCH}/os/Packages/centos-stream-release-8.6-1.el8.noarch.rpm \
        http://mirror.centos.org/centos/8-stream/BaseOS/${ARCH}/os/Packages/centos-stream-repos-8-6.el8.noarch.rpm \
        http://mirror.centos.org/centos/8-stream/BaseOS/${ARCH}/os/Packages/centos-gpg-keys-8-6.el8.noarch.rpm && \
      dnf config-manager --setopt=appstream*.exclude=*httpd*,mod_ssl --save \
    ; fi && \
    dnf -y install \
      httpd \
      mod_ssl \
      https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm \
      https://rpm.manageiq.org/release/18-radjabov/el8/noarch/manageiq-release-18.0-1.el8.noarch.rpm && \
    dnf -y update && \
    dnf -y module enable postgresql:13 ruby:3.1 && \
    if [[ "$RELEASE_BUILD" != "true" ]]; then dnf config-manager --enable manageiq-18-radjabov-nightly; fi && \
    dnf config-manager --setopt=ubi-8-*.exclude=dracut*,net-snmp*,redhat-release* --save && \
    if [[ "$LOCAL_RPM" = "true" ]]; then /create_local_yum_repo.sh; fi && \
    dnf -y install \
      ${RPM_PREFIX}-pods \
      python3-devel && \
    clean_dnf_rpm && \
    chgrp -R 0 $APP_ROOT && \
    chmod -R g=u $APP_ROOT

# Add in the container_env file now that the APP_ROOT is created from the RPM
ADD container-assets/container_env ${APP_ROOT}/

# Build the RPM manifest
RUN source /etc/default/evm && \
    /usr/bin/generate_rpm_manifest.sh && \
    clean_dnf_rpm

ENTRYPOINT ["/usr/bin/dumb-init", "--single-child", "--"]
