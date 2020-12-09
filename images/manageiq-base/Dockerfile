FROM registry.access.redhat.com/ubi8/ubi:8.3

ARG BUILD_REF=master
ARG BUILD_ORG=ManageIQ

ARG CORE_REPO_NAME=manageiq

ARG GIT_HOST=github.com
ARG GIT_AUTH

ARG HOOKS_SCRIPT_URL

RUN dnf -y --disableplugin=subscription-manager --setopt=tsflags=nodocs install git

RUN mkdir build && \
    if [[ -n "$GIT_AUTH" ]]; then GIT_HOST=${GIT_AUTH}@${GIT_HOST}; fi && curl -L https://${GIT_HOST}/${BUILD_ORG}/${CORE_REPO_NAME}-appliance-build/tarball/${BUILD_REF} | tar vxz -C build --strip 1

# Hooks should make modifications to the source in the build directory
RUN if [[ -n "$HOOKS_SCRIPT_URL" ]]; then curl -L ${HOOKS_SCRIPT_URL} | bash; fi

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
      https://rpm.manageiq.org/release/12-lasker/el8/noarch/manageiq-release-12.0-1.el8.noarch.rpm \
      https://resources.ovirt.org/pub/yum-repo/ovirt-release44.rpm && \
    dnf -y --disableplugin=subscription-manager module enable nodejs:12 && \
    dnf -y --disableplugin=subscription-manager module enable ruby:2.6 && \
    if [[ "$RELEASE_BUILD" != "true" ]]; then dnf config-manager --enable manageiq-12-lasker-nightly; fi && \
    dnf config-manager --setopt=ubi-8-*.exclude=redhat-release* --save && \
    if [[ "$LOCAL_RPM" = "true" ]]; then /create_local_yum_repo.sh; fi && \
    dnf -y --disableplugin=subscription-manager --setopt=tsflags=nodocs install \
      ${RPM_PREFIX}-pods          \
      python3-devel               \
      &&                          \
    dnf -y --disableplugin=subscription-manager upgrade && \
    dnf clean all

# Install python packages the same way the appliance does
COPY --from=0 build/kickstarts/partials/post/python_modules.ks.erb /tmp/python_modules
RUN bash /tmp/python_modules && rm -f /tmp/python_modules

RUN chgrp -R 0 $APP_ROOT && \
    chmod -R g=u $APP_ROOT

ADD container-assets/container_env ${APP_ROOT}

RUN curl -L -o /usr/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_${ARCH} && \
    chmod +x /usr/bin/dumb-init

RUN source /etc/default/evm && /usr/bin/generate_rpm_manifest.sh

ENTRYPOINT ["/usr/bin/dumb-init", "--single-child", "--"]
