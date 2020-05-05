FROM registry.access.redhat.com/ubi8/ubi:8.1

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

FROM registry.access.redhat.com/ubi8/ubi:8.1
MAINTAINER ManageIQ https://manageiq.org

ARG ARCH=x86_64

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

RUN curl -L https://copr.fedorainfracloud.org/coprs/manageiq/ManageIQ-Master/repo/epel-8/manageiq-ManageIQ-Master-epel-8.repo > /etc/yum.repos.d/manageiq.repo && \
    curl -L https://releases.ansible.com/ansible-runner/ansible-runner.el8.repo > /etc/yum.repos.d/ansible-runner.repo

RUN dnf -y --disableplugin=subscription-manager install http://mirror.centos.org/centos/8/BaseOS/${ARCH}/os/Packages/centos-repos-8.1-1.1911.0.8.el8.${ARCH}.rpm http://mirror.centos.org/centos/8/BaseOS/${ARCH}/os/Packages/centos-gpg-keys-8.1-1.1911.0.8.el8.noarch.rpm https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm && \
    dnf -y --disableplugin=subscription-manager module enable nodejs:12 && \
    dnf -y --disableplugin=subscription-manager module enable ruby:2.6 && \
    dnf config-manager --setopt=ubi-8-appstream.exclude=*net-snmp* --setopt=ubi-8-baseos.exclude=*net-snmp* --setopt=ubi-8-codeready-builder.exclude=*net-snmp* --save && \
    dnf -y --disableplugin=subscription-manager --setopt=tsflags=nodocs install \
      cmake                       \
      file                        \
      gcc-c++                     \
      git                         \
      glibc-langpack-en           \
      hostname                    \
      http-parser                 \
      iproute                     \
      libcurl-devel               \
      libpq                       \
      libssh2-devel               \
      libtool                     \
      libxml2-devel               \
      libxslt-devel               \
      lvm2                        \
      make                        \
      openssl-devel               \
      patch                       \
      postfix                     \
      postgresql-devel            \
      procps-ng                   \
      python3-devel               \
      redhat-rpm-config           \
      ruby-devel                  \
      sqlite-devel                \
      wget                        \
      xz                          \
      &&                          \
    dnf clean all

# Install python packages the same way the appliance does
COPY --from=0 build/kickstarts/partials/post/python_modules.ks.erb /tmp/python_modules
RUN bash /tmp/python_modules && rm -f /tmp/python_modules
COPY --from=0 build/kickstarts/partials/post/ovirt_ansible.ks.erb /tmp/ovirt_ansible
RUN bash /tmp/ovirt_ansible && rm -f /tmp/ovirt_ansible

RUN chgrp -R 0 $APP_ROOT && \
    chmod -R g=u $APP_ROOT

ADD container-assets/container_env ${APP_ROOT}

RUN wget -O /usr/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_${ARCH} && \
    chmod +x /usr/bin/dumb-init

ENTRYPOINT ["/usr/bin/dumb-init", "--single-child", "--"]
