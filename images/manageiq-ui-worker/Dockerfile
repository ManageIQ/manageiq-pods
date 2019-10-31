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
    npm install -g yarn && \
    rake update:ui && \
    RAILS_ENV=production rake evm:compile_assets && \
    rake evm:compile_sti_loader && \
    bin/rails log:clear tmp:clear && \
    npm cache clean --force && \
    rm -rvf ${APP_ROOT}/tmp/cache/assets && \
    rm -vf ${APP_ROOT}/log/*.log

# Build service UI
RUN source /etc/default/evm && \
    cd ${SUI_ROOT} && \
    yarn install && \
    yarn run available-languages && \
    yarn run build && \
    yarn cache clean

COPY container-assets/manageiq-http.conf /etc/httpd/conf.d
