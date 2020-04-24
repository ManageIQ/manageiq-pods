ARG FROM_REPO=manageiq
ARG FROM_TAG=latest-jansa

FROM ${FROM_REPO}/manageiq-webserver-worker:${FROM_TAG}
MAINTAINER ManageIQ https://manageiq.org

ARG NPM_REGISTRY_OVERRIDE

LABEL name="manageiq-ui-worker" \
      summary="ManageIQ user interface worker image"

# Install httpd and remove the existing httpd config from manageiq-appliance
RUN dnf -y --disableplugin=subscription-manager --setopt=tsflags=nodocs install httpd && \
    rm -f /etc/httpd/conf.d/* && \
    dnf clean all

COPY container-assets/npm_registry /tmp/npm_registry

# Compile assets for ManageIQ UI
RUN source /etc/default/evm && \
    export RAILS_USE_MEMORY_STORE="true" && \
    if [ ! -z "${NPM_REGISTRY_OVERRIDE}" ]; then \
      /tmp/npm_registry/npm_registry_setup.sh; \
    fi && \
    npm install -g yarn && \
    if [ ! -z "${NPM_REGISTRY_OVERRIDE}" ]; then \
      /tmp/npm_registry/yarn_registry_setup.sh; \
    fi && \
    rake update:ui && \
    RAILS_ENV=production rake evm:compile_assets && \
    rake evm:compile_sti_loader && \
    bin/rails log:clear tmp:clear && \
    npm cache clean --force && \
    rm -rvf ${APP_ROOT}/tmp/cache/assets && \
    rm -vf ${APP_ROOT}/log/*.log

# Build service UI
# PhantomJS is required only for testing and can be by-passed here
RUN source /etc/default/evm && \
    cd ${SUI_ROOT} && \
    sed -i '/phantomjs/d' package.json && \
    yarn install && \
    yarn run available-languages && \
    yarn run build && \
    rm -vf ${APP_ROOT}/log/*.log

RUN source /etc/default/evm && \
    yarn cache clean && \
    cd ${APP_ROOT} && \
    if [ ! -z "${NPM_REGISTRY_OVERRIDE}" ]; then \
      /tmp/npm_registry/npm_yarn_registry_cleanup.sh; \
    fi && \
    rm -rf /tmp/npm_registry && \
    rm -vf ${APP_ROOT}/log/*.log


# Configure httpd to run without root privileges
RUN chgrp root /var/run/httpd && chmod g+rwx /var/run/httpd && \
    chgrp root /var/log/httpd && chmod g+rwx /var/log/httpd
RUN sed -i '/^Listen 80/d' /etc/httpd/conf/httpd.conf

COPY container-assets/manageiq-http.conf /etc/httpd/conf.d
