FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/manageiq-operator \
    USER_UID=1001 \
    USER_NAME=manageiq-operator

# install operator binary
COPY build/_output/bin/manageiq-operator ${OPERATOR}
# install operator manifest
RUN mkdir -p /opt/manageiq/manifest
COPY build/_output/BUILD /opt/manageiq/manifest

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
