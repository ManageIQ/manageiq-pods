#!/bin/bash

yum -y install createrepo_c
rm -rf /tmp/rpms/repodata
createrepo /tmp/rpms
yum -y remove createrepo_c

ls -al /tmp/rpms

cat > /etc/yum.repos.d/local_rpm.repo << EOF
[local-rpm]
baseurl=file:///tmp/rpms/$basearch
name=Local yum repo
enabled=1
gpgcheck=0
EOF

dnf config-manager --setopt=manageiq-*.exclude=manageiq-* --save
