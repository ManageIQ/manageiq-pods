FROM registry.access.redhat.com/ubi8/ubi:latest as appliance_build

ARG BUILD_REF=najdorf
ARG BUILD_ORG=ManageIQ
ARG CORE_REPO_NAME=manageiq
ARG GIT_HOST=github.com
ARG GIT_AUTH

RUN mkdir build && \
    if [[ -n "$GIT_AUTH" ]]; then GIT_HOST=${GIT_AUTH}@${GIT_HOST}; fi && curl -L https://${GIT_HOST}/${BUILD_ORG}/${CORE_REPO_NAME}-appliance-build/tarball/${BUILD_REF} | tar vxz -C build --strip 1

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

RUN chmod -R g+w /etc/pki/ca-trust && \
    chmod -R g+w /usr/share/pki/ca-trust-legacy

COPY rpms/* /tmp/rpms/
COPY container-assets/create_local_yum_repo.sh /

RUN curl -L https://releases.ansible.com/ansible-runner/ansible-runner.el8.repo > /etc/yum.repos.d/ansible-runner.repo

RUN dnf -y --disableplugin=subscription-manager --setopt=tsflags=nodocs install \
      httpd \
      mod_ssl && \
    dnf -y remove *subscription-manager* && \
    if [ ${ARCH} != "s390x" ] ; then \
      dnf -y install \
        http://mirror.centos.org/centos/8-stream/BaseOS/${ARCH}/os/Packages/centos-stream-repos-8-2.el8.noarch.rpm \
        http://mirror.centos.org/centos/8-stream/BaseOS/${ARCH}/os/Packages/centos-gpg-keys-8-2.el8.noarch.rpm && \
      dnf config-manager --setopt=appstream*.exclude=*httpd* --save \
    ; fi && \
    dnf -y install \
      https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm \
      https://rpm.manageiq.org/release/14-najdorf/el8/noarch/manageiq-release-14.0-1.el8.noarch.rpm && \
    dnf -y module enable nodejs:14 && \
    dnf -y module enable ruby:2.7 && \
    if [[ "$RELEASE_BUILD" != "true" ]]; then dnf config-manager --enable manageiq-14-najdorf-nightly; fi && \
    dnf config-manager --setopt=ubi-8-*.exclude=dracut*,net-snmp*,perl-*,redhat-release* --save && \
    if [[ "$LOCAL_RPM" = "true" ]]; then /create_local_yum_repo.sh; fi && \
    dnf -y --setopt=tsflags=nodocs install \
      ${RPM_PREFIX}-pods          \
      python3-devel               \
      &&                          \
    dnf clean all && \
    rm -rf /var/cache/dnf && \
    chgrp -R 0 $APP_ROOT && \
    chmod -R g=u $APP_ROOT

# Install python packages the same way the appliance does
COPY --from=appliance_build build/kickstarts/partials/post/python_modules.ks.erb /tmp/python_modules
RUN bash /tmp/python_modules && \
    rm -f /tmp/python_modules && \
    rm -rf /root/.cache/pip && \
    dnf clean all && \
    rm -rf /var/cache/dnf

ADD container-assets/container_env ${APP_ROOT}

RUN curl -L -o /usr/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_${ARCH} && \
    chmod +x /usr/bin/dumb-init

RUN source /etc/default/evm && \
    /usr/bin/generate_rpm_manifest.sh && \
    dnf clean all && \
    rm -rf /var/cache/dnf

ENTRYPOINT ["/usr/bin/dumb-init", "--single-child", "--"]
