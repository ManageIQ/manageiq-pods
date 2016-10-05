FROM centos:centos7

# Memcached image for OpenShift ManageIQ

MAINTAINER ManageIQ https://github.com/ManageIQ/manageiq-appliance-build

LABEL io.k8s.description="Memcached is a general-purpose distributed memory object caching system" \
      io.k8s.display-name="Memcached" \
      io.openshift.expose-services="11211:memcached" \
      io.openshift.tags="memcached"

EXPOSE 11211

# Install latest memcached for Centos7

RUN yum install -y memcached && \
    yum -y --setopt=tsflags=nodocs install memcached && \
    rpm -V memcached && \
    yum clean all

COPY docker-assets/container-entrypoint /usr/bin

USER memcached
ENTRYPOINT ["container-entrypoint"]
CMD ["memcached"]
