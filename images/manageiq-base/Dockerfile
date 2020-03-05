FROM registry.access.redhat.com/ubi8/ubi

ARG MIQ_REF=master
ARG SUI_REF=master
ARG APPLIANCE_REF=master

ARG MIQ_ORG=ManageIQ
ARG SUI_ORG=ManageIQ
ARG APPLIANCE_ORG=ManageIQ

ARG CORE_REPO_NAME=manageiq

ARG GIT_HOST=github.com
ARG GIT_AUTH

RUN dnf -y --disableplugin=subscription-manager --setopt=tsflags=nodocs install git

RUN mkdir appliance && \
    if [[ -n "$GIT_AUTH" ]]; then GIT_HOST=${GIT_AUTH}@${GIT_HOST}; fi && curl -L https://${GIT_HOST}/${APPLIANCE_ORG}/${CORE_REPO_NAME}-appliance/tarball/${APPLIANCE_REF} | tar vxz -C appliance --strip 1

RUN mkdir sui && \
    if [[ -n "$GIT_AUTH" ]]; then GIT_HOST=${GIT_AUTH}@${GIT_HOST}; fi && curl -L https://${GIT_HOST}/${SUI_ORG}/${CORE_REPO_NAME}-ui-service/tarball/${SUI_REF} | tar vxz -C sui --strip 1

RUN mkdir app && \
    if [[ -n "$GIT_AUTH" ]]; then GIT_HOST=${GIT_AUTH}@${GIT_HOST}; fi && curl -L https://${GIT_HOST}/${MIQ_ORG}/${CORE_REPO_NAME}/tarball/${MIQ_REF} | tar vxz -C app --strip 1 && \
    echo "`date +'%Y%m%d%H%M%S'`_`git ls-remote https://${GIT_HOST}/${MIQ_ORG}/${CORE_REPO_NAME}.git ${MIQ_REF} | cut -c 1-7`" > app/BUILD

FROM registry.access.redhat.com/ubi8/ubi
MAINTAINER ManageIQ https://manageiq.org

ARG ARCH=x86_64

ENV TERM=xterm \
    CONTAINER=true \
    APP_ROOT=/var/www/miq/vmdb \
    APPLIANCE_ROOT=/opt/manageiq/manageiq-appliance \
    SUI_ROOT=/opt/manageiq/manageiq-ui-service

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
    dnf -y --disableplugin=subscription-manager --setopt=tsflags=nodocs install \
      ansible           \
      ansible-runner    \
      chrony            \
      cmake             \
      cronie            \
      file              \
      gcc-c++           \
      git               \
      glibc-langpack-en \
      hostname          \
      http-parser       \
      initscripts       \
      iproute           \
      libcurl-devel     \
      libpq             \
      libtool           \
      libxml2-devel     \
      libxslt-devel     \
      logrotate         \
      lvm2              \
      make              \
      net-tools         \
      nmap-ncat         \
      nodejs            \
      openldap-clients  \
      openscap-scanner  \
      openssl-devel     \
      patch             \
      psmisc            \
      postgresql-devel  \
      redhat-rpm-config \
      ruby-devel        \
      sqlite-devel      \
      v2v-conversion-host-ansible \
      wget              \
      which             \
      xz                \
      &&                \
    dnf clean all


RUN mkdir -p ${APP_ROOT} ${APPLIANCE_ROOT} ${SUI_ROOT} && \
    ln -vs ${APP_ROOT} /opt/manageiq/manageiq
COPY --from=0 app ${APP_ROOT}
COPY --from=0 appliance ${APPLIANCE_ROOT}
COPY --from=0 sui ${SUI_ROOT}

RUN ${APPLIANCE_ROOT}/setup && \
    mkdir -p ${APP_ROOT}/log/apache && \
    rm -f /etc/httpd/conf.d/* && \
    cp ${APP_ROOT}/config/cable.yml.sample ${APP_ROOT}/config/cable.yml

# build and install gems
WORKDIR ${APP_ROOT}
RUN source /etc/default/evm && \
    export RAILS_USE_MEMORY_STORE="true" && \
    gem install bundler -v ">=1.16.2" && \
    if [ ${ARCH} = "ppc64le" ] ; then gem install sassc  -- --disable-march-tune-native ; fi && \
    bundle install && \
    rake tmp:clear evm:compile_sti_loader log:clear && \
    find ${RUBY_GEMS_ROOT}/gems/ -name .git | xargs rm -rvf && \
    find ${RUBY_GEMS_ROOT}/gems/ | grep "\.o$" | xargs rm -rvf && \
    rm -rvf ${RUBY_GEMS_ROOT}/gems/rugged-*/vendor/libgit2/build && \
    rm -rvf ${RUBY_GEMS_ROOT}/cache/* && \
    rm -rvf /root/.bundle/cache && \
    rm -rvf ${APP_ROOT}/tmp/cache/assets && \
    rm -vf ${APP_ROOT}/log/*.log

ADD container-assets/container_env ${APP_ROOT}

RUN wget -O /usr/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_${ARCH} && \
    chmod +x /usr/bin/dumb-init

ENTRYPOINT ["/usr/bin/dumb-init", "--single-child", "--"]
