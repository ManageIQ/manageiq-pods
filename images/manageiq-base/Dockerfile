FROM registry.access.redhat.com/ubi8/ubi:8.3 as appliance_build

ARG BUILD_REF=master
ARG BUILD_ORG=ManageIQ
ARG CORE_REPO_NAME=manageiq
ARG GIT_HOST=github.com
ARG GIT_AUTH

RUN mkdir build && \
    if [[ -n "$GIT_AUTH" ]]; then GIT_HOST=${GIT_AUTH}@${GIT_HOST}; fi && curl -L https://${GIT_HOST}/${BUILD_ORG}/${CORE_REPO_NAME}-appliance-build/tarball/${BUILD_REF} | tar vxz -C build --strip 1

FROM registry.access.redhat.com/ubi8/ubi:8.3
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

COPY rpms/* /tmp/rpms/
COPY container-assets/create_local_yum_repo.sh /

RUN curl -L https://releases.ansible.com/ansible-runner/ansible-runner.el8.repo > /etc/yum.repos.d/ansible-runner.repo

RUN dnf -y --disableplugin=subscription-manager install \
      http://mirror.centos.org/centos/8.3.2011/BaseOS/${ARCH}/os/Packages/centos-linux-repos-8-2.el8.noarch.rpm \
      http://mirror.centos.org/centos/8.3.2011/BaseOS/${ARCH}/os/Packages/centos-gpg-keys-8-2.el8.noarch.rpm \
      https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm \
      https://rpm.manageiq.org/release/13-morphy/el8/noarch/manageiq-release-13.0-1.el8.noarch.rpm \
      https://resources.ovirt.org/pub/yum-repo/ovirt-release44.rpm && \
    dnf -y --disableplugin=subscription-manager module enable nodejs:12 && \
    dnf -y --disableplugin=subscription-manager module enable ruby:2.6 && \
    if [[ "$RELEASE_BUILD" != "true" ]]; then dnf config-manager --enable manageiq-13-morphy-nightly; fi && \
    dnf config-manager --setopt=ubi-8-*.exclude=dracut*,net-snmp*,redhat-release*,rpm* --save && \
    if [[ "$LOCAL_RPM" = "true" ]]; then /create_local_yum_repo.sh; fi && \
    dnf -y --disableplugin=subscription-manager --setopt=tsflags=nodocs install \
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
