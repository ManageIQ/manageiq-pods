FROM manageiq/manageiq-pods:backend-latest
MAINTAINER ManageIQ https://github.com/ManageIQ/manageiq-appliance-build

## Set build ARGs
ARG REF=master
ARG GHORG=ManageIQ

## Set ENV, LANG only needed if building with docker-1.8
ENV SUI_ROOT=/opt/manageiq/manageiq-ui-service

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
      io.openshift.expose-services="80:http" \
      io.openshift.tags="ManageIQ,miq,manageiq"

## Install EPEL repo, yum necessary packages for the build without docs, clean all caches
RUN yum -y install centos-release-scl-rh && \
    yum -y install httpd --setopt=tsflags=nodocs && \
    yum clean all

## GIT clone service UI repo (SUI)
RUN mkdir -p ${SUI_ROOT} && \
    curl -L https://github.com/${GHORG}/manageiq-ui-service/tarball/${REF} | tar vxz -C ${SUI_ROOT} --strip 1

## Setup environment
RUN rm -f /etc/httpd/conf.d/ssl.conf && \
    rm -f /etc/httpd/conf.d/manageiq-http*.conf

## Change workdir to application root, build/install gems
WORKDIR ${APP_ROOT}
RUN source /etc/default/evm && \
    export RAILS_USE_MEMORY_STORE="true" && \
    rake update:ui && \
    bin/rails log:clear tmp:clear && \
    rake evm:compile_assets && \
    # Cleanup install artifacts
    npm cache clean && \
    bower cache clean && \
    rm -rvf ${APP_ROOT}/tmp/cache/assets && \
    rm -vf ${APP_ROOT}/log/*.log

## Build SUI
RUN source /etc/default/evm && \
    cd ${SUI_ROOT} && \
    yarn install --production && \
    yarn run build && \
    yarn cache clean

## Expose required container ports
EXPOSE 80

COPY docker-assets/check-dependent-services.sh /bin
COPY docker-assets/manageiq-http.conf /etc/httpd/conf.d

ENTRYPOINT ["/usr/local/bin/dumb-init", "--single-child", "--"]
CMD ["entrypoint"]
