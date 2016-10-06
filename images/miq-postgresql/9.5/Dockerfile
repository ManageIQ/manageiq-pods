# This image is layered from SCL centos7 openshift postgresql 9.5
# https://github.com/sclorg/postgresql-container/tree/master/9.5
FROM centos/postgresql-95-centos7

MAINTAINER ManageIQ https://github.com/ManageIQ/manageiq-appliance-build

# Switch USER to root to add required repo and packages
USER root

# Fetch postgresql 9.5 COPR and pglogical repos for MIQ use
RUN  curl -sSLko /etc/yum.repos.d/ncarboni-pglogical-SCL-epel-7.repo \
     https://copr.fedorainfracloud.org/coprs/ncarboni/pglogical-SCL/repo/epel-7/ncarboni-pglogical-SCL-epel-7.repo

RUN yum -y --setopt=tsflags=nodocs install rh-postgresql95-postgresql-pglogical-output rh-postgresql95-postgresql-pglogical rh-repmgr95 && \
    yum clean all

# Add pglogical openshift tag to new image
LABEL io.openshift.tags="database,postgresql,postgresql95,rh-postgresql95,pglogical"

# Modified pg startup script to add required role
COPY docker-assets/run-postgresql /usr/bin

# Loosen permission bits to avoid problems running container with arbitrary UID
RUN /usr/libexec/fix-permissions /var/lib/pgsql && \
    /usr/libexec/fix-permissions /var/run/postgresql

# Switch USER back to postgres
USER 26
