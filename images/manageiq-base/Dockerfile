FROM manageiq/ruby:2.5
MAINTAINER ManageIQ https://manageiq.org

ARG MIQ_REF=master
ARG SUI_REF=master
ARG APPLIANCE_REF=master

ARG MIQ_ORG=ManageIQ
ARG SUI_ORG=ManageIQ
ARG APPLIANCE_ORG=ManageIQ

ARG CORE_REPO_NAME=manageiq

ARG GIT_HOST=github.com
ARG GIT_AUTH_AVAILABLE

ENV TERM=xterm \
    CONTAINER=true \
    APP_ROOT=/var/www/miq/vmdb \
    APPLIANCE_ROOT=/opt/manageiq/manageiq-appliance \
    SUI_ROOT=/opt/manageiq/manageiq-ui-service \
    IMAGE_VERSION=${MIQ_REF}

LABEL name="manageiq-base" \
      vendor="ManageIQ" \
      version="Master" \
      release=${MIQ_REF} \
      url="https://manageiq.org/" \
      summary="ManageIQ base application image" \
      description="ManageIQ is a management and automation platform for virtual, private, and hybrid cloud infrastructures." \
      io.k8s.display-name="ManageIQ" \
      io.k8s.description="ManageIQ is a management and automation platform for virtual, private, and hybrid cloud infrastructures." \
      io.openshift.tags="ManageIQ,miq,manageiq"

RUN yum -y install https://rpm.nodesource.com/pub_10.x/el/7/x86_64/nodesource-release-el7-1.noarch.rpm && \
    yum -y install --setopt=tsflags=nodocs \
                   chrony           \
                   cmake            \
                   cronie           \
                   file             \
                   gcc-c++          \
                   git              \
                   http-parser      \
                   initscripts      \
                   libcurl-devel    \
                   libpq            \
                   libtool          \
                   libxslt-devel    \
                   logrotate        \
                   lvm2             \
                   net-tools        \
                   nmap-ncat        \
                   nodejs           \
                   openldap-clients \
                   openscap-scanner \
                   patch            \
                   psmisc           \
                   postgresql-devel \
                   sqlite-devel     \
                   sysvinit-tools   \
                   which            \
                   &&               \
    yum clean all

RUN if [[ -n "$GIT_AUTH_AVAILABLE" ]] ; then \
      export GIT_AUTH=$(wget -O - -q http://localhost:8000/git_auth.txt) ; \
      export GIT_HOST=${GIT_AUTH}@${GIT_HOST} ; \
    fi && \
    mkdir -p ${APPLIANCE_ROOT} && \
    curl -L https://${GIT_HOST}/${APPLIANCE_ORG}/${CORE_REPO_NAME}-appliance/tarball/${APPLIANCE_REF} | tar vxz -C ${APPLIANCE_ROOT} --strip 1

RUN if [[ -n "$GIT_AUTH_AVAILABLE" ]] ; then \
      export GIT_AUTH=$(wget -O - -q http://localhost:8000/git_auth.txt) ; \
      export GIT_HOST=${GIT_AUTH}@${GIT_HOST} ; \
    fi && \
    mkdir -p ${SUI_ROOT} && \
    curl -L https://${GIT_HOST}/${SUI_ORG}/${CORE_REPO_NAME}-ui-service/tarball/${SUI_REF} | tar vxz -C ${SUI_ROOT} --strip 1

RUN if [[ -n "$GIT_AUTH_AVAILABLE" ]] ; then \
      export GIT_AUTH=$(wget -O - -q http://localhost:8000/git_auth.txt) ; \
      export GIT_HOST=${GIT_AUTH}@${GIT_HOST} ; \
    fi && \
    mkdir -p ${APP_ROOT} && \
    ln -vs ${APP_ROOT} /opt/manageiq/manageiq && \
    curl -L https://${GIT_HOST}/${MIQ_ORG}/${CORE_REPO_NAME}/tarball/${MIQ_REF} | tar vxz -C ${APP_ROOT} --strip 1 && \
    echo "`date +'%Y%m%d%H%M%S'`_`git ls-remote https://${GIT_HOST}/${MIQ_ORG}/${CORE_REPO_NAME}.git ${MIQ_REF} | cut -c 1-7`" > ${APP_ROOT}/BUILD

RUN ${APPLIANCE_ROOT}/setup && \
    mkdir -p ${APP_ROOT}/log/apache && \
    rm -f /etc/httpd/conf.d/* && \
    cp ${APP_ROOT}/config/cable.yml.sample ${APP_ROOT}/config/cable.yml

# build and install gems
WORKDIR ${APP_ROOT}
RUN source /etc/default/evm && \
    export RAILS_USE_MEMORY_STORE="true" && \
    gem install bundler -v ">=1.16.2" && \
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

RUN wget -O /usr/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.0/dumb-init_1.2.0_amd64 && \
    chmod +x /usr/bin/dumb-init

ENTRYPOINT ["/usr/bin/dumb-init", "--single-child", "--"]
