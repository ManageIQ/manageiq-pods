FROM manageiq/ruby:2.4
MAINTAINER ManageIQ https://github.com/ManageIQ/manageiq-appliance-build

## Set build ARGs
ARG REF=master

## Set ENV, LANG only needed if building with docker-1.8
ENV TERM=xterm \
    CONTAINER=true \
    APP_ROOT=/var/www/miq/vmdb \
    APP_ROOT_PERSISTENT=/persistent \
    APPLIANCE_ROOT=/opt/manageiq/manageiq-appliance \
    CONTAINER_SCRIPTS_ROOT=/opt/manageiq/container-scripts \
    IMAGE_VERSION=${REF}

## Atomic/OpenShift Labels
LABEL name="manageiq" \
      vendor="ManageIQ" \
      version="Master" \
      release=${REF} \
      url="http://manageiq.org/" \
      summary="ManageIQ appliance image" \
      description="ManageIQ is a management and automation platform for virtual, private, and hybrid cloud infrastructures." \
      io.k8s.display-name="ManageIQ" \
      io.k8s.description="ManageIQ is a management and automation platform for virtual, private, and hybrid cloud infrastructures." \
      io.openshift.tags="ManageIQ,miq,manageiq"

# Fetch MIQ repo for http-parser
RUN curl -sSLko /etc/yum.repos.d/manageiq-ManageIQ-Master-epel-7.repo \
      https://copr.fedorainfracloud.org/coprs/manageiq/ManageIQ-Master/repo/epel-7/manageiq-ManageIQ-Master-epel-7.repo

## Install EPEL repo, yum necessary packages for the build without docs, clean all caches
RUN yum -y install centos-release-scl-rh \
                   https://rpm.nodesource.com/pub_8.x/el/7/x86_64/nodesource-release-el7-1.noarch.rpm && \
    yum -y install --setopt=tsflags=nodocs \
                   chrony                  \
                   cmake                   \
                   cronie                  \
                   file                    \
                   gcc-c++                 \
                   git                     \
                   http-parser             \
                   initscripts             \
                   libcurl-devel           \
                   libtool                 \
                   libxslt-devel           \
                   logrotate               \
                   lvm2                    \
                   net-tools               \
                   nmap-ncat               \
                   nodejs                  \
                   openldap-clients        \
                   openscap-scanner        \
                   patch                   \
                   psmisc                  \
                   rh-postgresql95-postgresql-devel  \
                   rh-postgresql95-postgresql-libs \
                   sqlite-devel            \
                   sysvinit-tools          \
                   which                   \
                   &&                      \
    yum clean all

## GIT clone manageiq-appliance
RUN mkdir -p ${APPLIANCE_ROOT} && \
    curl -L https://github.com/ManageIQ/manageiq-appliance/tarball/${REF} | tar vxz -C ${APPLIANCE_ROOT} --strip 1

## GIT clone manageiq
RUN mkdir -p ${APP_ROOT} && \
    ln -vs ${APP_ROOT} /opt/manageiq/manageiq && \
    curl -L https://github.com/ManageIQ/manageiq/tarball/${REF} | tar vxz -C ${APP_ROOT} --strip 1 && \
    echo "`date +'%Y%m%d%H%M%S'`_`git ls-remote https://github.com/ManageIQ/manageiq.git ${REF} | cut -c 1-7`" > ${APP_ROOT}/BUILD

## Setup environment
RUN ${APPLIANCE_ROOT}/setup && \
    mkdir -p ${APP_ROOT}/log/apache && \
    mkdir ${APP_ROOT_PERSISTENT} && \
    mkdir -p ${CONTAINER_SCRIPTS_ROOT} && \
    cp ${APP_ROOT}/config/cable.yml.sample ${APP_ROOT}/config/cable.yml

## Change workdir to application root, build/install gems
WORKDIR ${APP_ROOT}
RUN source /etc/default/evm && \
    export RAILS_USE_MEMORY_STORE="true" && \
    npm install bower yarn -g && \
    gem install bundler -v ">=1.16.2" && \
    bundle install && \
    rake update:ui && \
    bin/rails log:clear tmp:clear && \
    rake evm:compile_assets && \
    rake evm:compile_sti_loader && \
    # Cleanup install artifacts
    npm cache clean --force && \
    bower cache clean && \
    find ${RUBY_GEMS_ROOT}/gems/ -name .git | xargs rm -rvf && \
    find ${RUBY_GEMS_ROOT}/gems/ | grep "\.s\?o$" | xargs rm -rvf && \
    rm -rvf ${RUBY_GEMS_ROOT}/gems/rugged-*/vendor/libgit2/build && \
    rm -rvf ${RUBY_GEMS_ROOT}/cache/* && \
    rm -rvf /root/.bundle/cache && \
    rm -rvf ${APP_ROOT}/tmp/cache/assets && \
    rm -vf ${APP_ROOT}/log/*.log

## Copy OpenShift and appliance-initialize scripts
COPY container-assets/entrypoint /usr/bin
COPY container-assets/container.data.persist /
COPY container-assets/appliance-initialize.sh /bin
COPY container-assets/check-dependent-services.sh /bin
COPY container-assets/miq_logs.conf /etc/logrotate.d
ADD  container-assets/container-scripts ${CONTAINER_SCRIPTS_ROOT}

RUN wget -O /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.0/dumb-init_1.2.0_amd64 && \
    chmod +x /usr/local/bin/dumb-init

ENTRYPOINT ["/usr/local/bin/dumb-init", "--single-child", "--"]
CMD ["entrypoint"]
