ARG FROM_REPO=docker.io/manageiq
ARG FROM_TAG=latest

# alias stage for rpms
FROM ${FROM_REPO}/manageiq-rpms:${FROM_TAG} as rpms

################################################################################

FROM registry.access.redhat.com/ubi10/ubi
ARG FROM_REPO
ARG FROM_TAG
ARG RELEASE_BUILD
ARG RPM_PREFIX=manageiq

ENV TERM=xterm \
    CONTAINER=true \
    APP_ROOT=/var/www/miq/vmdb

LABEL name="manageiq-base" \
      summary="ManageIQ base application image" \
      vendor="ManageIQ" \
      url="https://manageiq.org/" \
      description="ManageIQ is a management and automation platform for virtual, private, and hybrid cloud infrastructures." \
      io.k8s.display-name="ManageIQ" \
      io.k8s.description="ManageIQ is a management and automation platform for virtual, private, and hybrid cloud infrastructures." \
      io.openshift.tags="ManageIQ,miq,manageiq"

# Force the sticky bit on /tmp - https://bugzilla.redhat.com/show_bug.cgi?id=2138434
RUN chmod +t /tmp

RUN chmod -R g+w /etc/pki/ca-trust && \
    chmod -R g+w /usr/share/pki/ca-trust-legacy

RUN --mount=type=bind,from=quay.io/manageiq/build_tools:v1,source=/tools,target=/usr/local/bin --mount=type=bind,from=rpms,source=/tmp/rpms,target=/tmp/rpms \
    dnf -y --setopt=tsflags=nodocs update && \
    miq_switch_to_centos_stream_10 && \
    miq_enable_epel && \
    dnf config-manager --setopt=tsflags=nodocs --setopt=install_weak_deps=False --save && \
    dnf -y install https://rpm.manageiq.org/release/21-uhlmann/el10/noarch/manageiq-release-21.0-1.el10.noarch.rpm && \
    if [[ "$RELEASE_BUILD" != "true" ]]; then dnf config-manager --enable manageiq-21-uhlmann-nightly; fi && \
    if [[ -n "$(ls /tmp/rpms)" ]]; then miq_prepare_local_yum_repo.sh; fi && \
    # Currently the minimum version of qpid-proton in EPEL is 0.40.0, but we're still on 0.37, so pull it from the ManageIQ RPM repo instead https://github.com/ManageIQ/manageiq/blob/b5bf7bee569991d3f1a182a4792fae8cb6265863/Gemfile#L220
    dnf config-manager --setopt=epel.exclude=*qpid-proton* --save && \
    dnf -y install \
      dumb-init \
      httpd \
      mod_ssl \
      ${RPM_PREFIX}-pods \
      python3-devel && \
    miq_clean_dnf_rpm && \
    chown manageiq:root /home/manageiq && \
    chmod -R g=u /home/manageiq && \
    chgrp -R 0 $APP_ROOT && \
    chmod -R g=u $APP_ROOT

# Add in the container_env file now that the APP_ROOT is created from the RPM
ADD container-assets/container_env ${APP_ROOT}/

# Build the RPM manifest
RUN --mount=type=bind,from=quay.io/manageiq/build_tools:v1,source=/tools,target=/usr/local/bin --mount=type=bind,from=rpms,source=/tmp/rpms,target=/tmp/rpms \
    source /etc/default/evm && \
    /usr/bin/generate_rpm_manifest.sh && \
    miq_clean_dnf_rpm

ENTRYPOINT ["/usr/bin/dumb-init", "--single-child", "--"]
