FROM quay.io/centos/centos:stream10

COPY rpms/* /tmp/rpms/
RUN  rm -f /tmp/rpms/.gitkeep

RUN if [[ -n "$(ls /tmp/rpms)" ]]; then \
      ls -al /tmp/rpms && \
      dnf -y install createrepo_c && \
      rm -rf /tmp/rpms/repodata && \
      createrepo /tmp/rpms; \
    fi
