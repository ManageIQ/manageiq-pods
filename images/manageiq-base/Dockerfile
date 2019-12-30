FROM manageiq/ruby:2.5
MAINTAINER ManageIQ https://manageiq.org

ARG MIQ_REF=master
ARG SUI_REF=master
ARG APPLIANCE_REF=master

ARG MIQ_ORG=ManageIQ
ARG SUI_ORG=ManageIQ
ARG APPLIANCE_ORG=ManageIQ

ARG ARCH=amd64

ENV TERM=xterm \
    CONTAINER=true \
    APP_ROOT=/var/www/miq/vmdb \
    APPLIANCE_ROOT=/opt/manageiq/manageiq-appliance \
    APPLIANCE_PG_DATA=/var/lib/pgsql/data \
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

RUN yum -y install --setopt=tsflags=nodocs \
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
                   openldap-clients \
                   openscap-scanner \
                   patch            \
                   psmisc           \
                   sqlite-devel     \
                   sysvinit-tools   \
                   which            \
		   wget 	    \
		   make		    \
		   readline-devel   \
		   zlib-devel	    \
		   perl-devel	    \
		   bison	    \
		   flex		    \
                   &&               \
    yum clean all

RUN wget  https://github.com/postgres/postgres/archive/REL_10_11.tar.gz && \
    tar -zxvf REL_10_11.tar.gz && cd postgres-REL_10_11/ && ./configure --prefix=/usr && make && make install && \
    adduser postgres && mkdir -p ${APPLIANCE_PG_DATA} && chown postgres ${APPLIANCE_PG_DATA}


# nvm environment variables
ENV NVM_DIR /usr/local/nvm
ENV NODE_VERSION 10.17.0
# install nvm
# https://github.com/creationix/nvm#install-script
RUN curl --silent -o- https://raw.githubusercontent.com/creationix/nvm/v0.31.2/install.sh | bash
# install node and npm
RUN source $NVM_DIR/nvm.sh \
    && nvm install $NODE_VERSION \
    && nvm alias default $NODE_VERSION \
    && nvm use default
# add node and npm to path so the commands are available
ENV NODE_PATH $NVM_DIR/v$NODE_VERSION/lib/node_modules
ENV PATH $NVM_DIR/versions/node/v$NODE_VERSION/bin:$PATH

RUN mkdir -p ${APPLIANCE_ROOT} && \
    curl -L https://github.com/${APPLIANCE_ORG}/manageiq-appliance/tarball/${APPLIANCE_REF} | tar vxz -C ${APPLIANCE_ROOT} --strip 1

RUN mkdir -p ${SUI_ROOT} && \
    curl -L https://github.com/${SUI_ORG}/manageiq-ui-service/tarball/${SUI_REF} | tar vxz -C ${SUI_ROOT} --strip 1

RUN mkdir -p ${APP_ROOT} && \
    ln -vs ${APP_ROOT} /opt/manageiq/manageiq && \
    curl -L https://github.com/${MIQ_ORG}/manageiq/tarball/${MIQ_REF} | tar vxz -C ${APP_ROOT} --strip 1 && \
    echo "`date +'%Y%m%d%H%M%S'`_`git ls-remote https://github.com/${MIQ_ORG}/manageiq.git ${MIQ_REF} | cut -c 1-7`" > ${APP_ROOT}/BUILD

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
