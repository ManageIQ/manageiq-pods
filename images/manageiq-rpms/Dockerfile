FROM registry.access.redhat.com/ubi9/ubi

COPY rpms/* /tmp/rpms/
RUN  rm -f /tmp/rpms/.gitkeep

RUN if [[ -n "$(ls /tmp/rpms)" ]]; then \
      ARCH=$(uname -m) && \
      sed -i 's/enabled=1/enabled=0/g' /etc/dnf/plugins/subscription-manager.conf && \
      dnf config-manager --setopt=tsflags=nodocs --setopt=install_weak_deps=False --save && \
      dnf -y --setopt=protected_packages= remove redhat-release && \
      dnf -y install \
        http://mirror.stream.centos.org/9-stream/BaseOS/${ARCH}/os/Packages/centos-stream-release-9.0-24.el9.noarch.rpm \
        http://mirror.stream.centos.org/9-stream/BaseOS/${ARCH}/os/Packages/centos-stream-repos-9.0-24.el9.noarch.rpm \
        http://mirror.stream.centos.org/9-stream/BaseOS/${ARCH}/os/Packages/centos-gpg-keys-9.0-24.el9.noarch.rpm ; \
    fi

RUN if [[ -n "$(ls /tmp/rpms)" ]]; then \
      ls -al /tmp/rpms && \
      yum -y install createrepo_c && \
      rm -rf /tmp/rpms/repodata && \
      createrepo /tmp/rpms; \
    fi
