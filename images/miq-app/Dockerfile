FROM manageiq/ruby
MAINTAINER ManageIQ https://github.com/ManageIQ/manageiq-appliance-build

## Set build ARGs
ARG REF=master

## Set ENV, LANG only needed if building with docker-1.8
ENV container=docker \
    TERM=xterm \
    CONTAINER=true \
    APP_ROOT=/var/www/miq/vmdb \
    APP_ROOT_PERSISTENT=/persistent \
    APP_ROOT_PERSISTENT_REGION=/persistent-region \
    APPLIANCE_ROOT=/opt/manageiq/manageiq-appliance \
    SUI_ROOT=/opt/manageiq/manageiq-ui-service \ 
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
      io.openshift.expose-services="443:https" \
      io.openshift.tags="ManageIQ,miq,manageiq"

## To cleanly shutdown systemd, use SIGRTMIN+3
STOPSIGNAL SIGRTMIN+3

## Install EPEL repo, yum necessary packages for the build without docs, clean all caches
RUN yum -y install centos-release-scl-rh && \
    yum -y install --setopt=tsflags=nodocs \
                   cmake                   \
                   file                    \
                   gcc-c++                 \
                   git                     \
                   libcurl-devel           \
                   libtool                 \
                   libxslt-devel           \
                   net-tools               \
                   nodejs                  \
                   openscap-scanner        \
                   patch                   \
                   rh-postgresql95-postgresql-libs \
                   rh-postgresql95-postgresql-devel  \
                   sqlite-devel            \
                   sysvinit-tools          \
                   which                   \
                   httpd                   \
                   mod_ssl                 \
                   mod_auth_kerb           \
                   mod_authnz_pam          \
                   mod_intercept_form_submit \
                   mod_lookup_identity     \
                   initscripts             \
                   npm                     \
                   chrony                  \
                   psmisc                  \
                   lvm2                    \
                   openldap-clients        \
                   cronie                  \
                   logrotate               \
                   nmap-ncat               \
                   &&                      \
    yum clean all

## Systemd cleanup base image
RUN (cd /lib/systemd/system/sysinit.target.wants && for i in *; do [ $i == systemd-tmpfiles-setup.service ] || rm -vf $i; done) && \
    rm -vf /lib/systemd/system/multi-user.target.wants/* && \
    rm -vf /etc/systemd/system/*.wants/* && \
    rm -vf /lib/systemd/system/local-fs.target.wants/* && \
    rm -vf /lib/systemd/system/sockets.target.wants/*udev* && \
    rm -vf /lib/systemd/system/sockets.target.wants/*initctl* && \
    rm -vf /lib/systemd/system/basic.target.wants/* && \
    rm -vf /lib/systemd/system/anaconda.target.wants/*

## GIT clone manageiq-appliance and service UI repo (SUI)
RUN mkdir -p ${APP_ROOT} && \
    mkdir -p ${APPLIANCE_ROOT} && \
    mkdir -p ${SUI_ROOT} && \
    ln -vs ${APP_ROOT} /opt/manageiq/manageiq && \
    curl -L https://github.com/ManageIQ/manageiq-appliance/tarball/${REF} | tar vxz -C ${APPLIANCE_ROOT} --strip 1 && \
    curl -L https://github.com/ManageIQ/manageiq-ui-service/tarball/${REF} | tar vxz -C ${SUI_ROOT} --strip 1

## Add ManageIQ source from local directory (dockerfile development) or from Github (official build)
#ADD . ${APP_ROOT}
RUN curl -L https://github.com/ManageIQ/manageiq/tarball/${REF} | tar vxz -C ${APP_ROOT} --strip 1

## Setup environment
RUN ${APPLIANCE_ROOT}/setup && \
    mkdir ${APP_ROOT}/log/apache && \
    mkdir ${APP_ROOT_PERSISTENT} && \
    mkdir ${APP_ROOT_PERSISTENT_REGION} && \
    mkdir -p ${CONTAINER_SCRIPTS_ROOT} && \
    mv /etc/httpd/conf.d/ssl.conf{,.orig} && \
    echo "# This file intentionally left blank. ManageIQ maintains its own SSL configuration" > /etc/httpd/conf.d/ssl.conf && \
    cp ${APP_ROOT}/config/cable.yml.sample ${APP_ROOT}/config/cable.yml

## Change workdir to application root, build/install gems
WORKDIR ${APP_ROOT}
RUN source /etc/default/evm && \
    export RAILS_USE_MEMORY_STORE="true" && \
    npm install bower yarn -g && \
    gem install bundler --conservative && \
    bundle install && \
    rake update:bower && \
    bin/rails log:clear tmp:clear && \
    rake evm:compile_assets && \
    rake evm:compile_sti_loader && \
    # Cleanup install artifacts
    npm cache clean && \
    bower cache clean && \
    find ${RUBY_GEMS_ROOT}/gems/ -name .git | xargs rm -rvf && \
    find ${RUBY_GEMS_ROOT}/gems/ | grep "\.s\?o$" | xargs rm -rvf && \
    rm -rvf ${RUBY_GEMS_ROOT}/gems/rugged-*/vendor/libgit2/build && \
    rm -rvf ${RUBY_GEMS_ROOT}/cache/* && \
    rm -rvf /root/.bundle/cache && \
    rm -rvf ${APP_ROOT}/tmp/cache/assets && \
    rm -vf ${APP_ROOT}/log/*.log

## Build SUI
RUN source /etc/default/evm && \
    cd ${SUI_ROOT} && \
    yarn install --production && \
    yarn run build && \
    yarn cache clean

## Expose required container ports
EXPOSE 80 443

## Copy OpenShift and appliance-initialize scripts
COPY docker-assets/entrypoint /usr/bin
COPY docker-assets/container.data.persist /
COPY docker-assets/appliance-initialize.sh /bin
ADD  docker-assets/container-scripts ${CONTAINER_SCRIPTS_ROOT}

RUN wget -O /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.0/dumb-init_1.2.0_amd64
RUN chmod +x /usr/local/bin/dumb-init

## Call systemd to bring up system
ENTRYPOINT ["/usr/local/bin/dumb-init", "--"]
CMD ["entrypoint"]
