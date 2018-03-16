FROM manageiq/manageiq-webserver-worker:latest
MAINTAINER ManageIQ https://manageiq.org

LABEL name="manageiq-ui-worker" \
      summary="ManageIQ user interface worker image"

# Install httpd and remove the existing httpd config from manageiq-appliance
RUN yum -y install httpd --setopt=tsflags=nodocs && \
    rm -f /etc/httpd/conf.d/* && \
    yum clean all

# Compile assets for ManageIQ UI
RUN source /etc/default/evm && \
    export RAILS_USE_MEMORY_STORE="true" && \
    npm install bower yarn -g && \
    rake update:ui && \
    rake evm:compile_assets && \
    bin/rails log:clear tmp:clear && \
    npm cache clean && \
    bower cache clean && \
    rm -rvf ${APP_ROOT}/tmp/cache/assets && \
    rm -vf ${APP_ROOT}/log/*.log

# Build service UI
RUN source /etc/default/evm && \
    cd ${SUI_ROOT} && \
    yarn install --production && \
    yarn run build && \
    yarn cache clean

COPY container-assets/manageiq-http.conf /etc/httpd/conf.d
