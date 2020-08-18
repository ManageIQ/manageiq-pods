#!/bin/bash

set -o pipefail
set -o errtrace
set -o nounset
set -o errexit

([[ ! -v CI_USER_TOKEN ]]  ||  [[ -z "$CI_USER_TOKEN" ]]) && {
  echo "GIT Personal Access Token required to run the stitchcontainer script. Define GIT"
  echo "Personal Access token as environment variable CI_USER_TOKEN to run and build the" 
  echo "container. "
  echo
  echo "To build manageiq container in Travis, make sure CI_USER_TOKEN environment "
  echo "variable is also declared and defined with Personal Access Token in Travis." 
  exit 1
}

MANAGEIQ_ADDR=https://$CI_USER_TOKEN@github.com/ManageIQ/manageiq.git
MANAGEIQ_RPM_ADDR=https://$CI_USER_TOKEN@github.com/ManageIQ/manageiq-rpm_build.git
MANAGEIQ_PODS_ADDR=https://$CI_USER_TOKEN@github.com/ManageIQ/manageiq-pods.git

# if the directory exist, then remove it
[ -d manageiq-container ] && rm -rf manageiq-container
mkdir -p manageiq-container
pushd manageiq-container


  # now git clone the manageiq-rpm_build
  git clone $MANAGEIQ_RPM_ADDR manageiq-rpm_build --depth 1 || {
    echo "Unable to clone manageiq-rpm_build repo" >&2
  }

  pushd manageiq-rpm_build
    mkdir -p OPTIONS
    cat > OPTIONS/options.yml << EndOfMessage
---
product_name: manageiq
repos:
  ref:        master
  manageiq:
    url:      $MANAGEIQ_ADDR
    ref:      master
EndOfMessage
    cat OPTIONS/options.yml

    mkdir -p OUTPUT
    docker build -t manageiq-rpm:latest .
    docker run -e CI_USER_TOKEN=$CI_USER_TOKEN -v $PWD/OPTIONS:/root/OPTIONS -v $PWD/OUTPUT:/root/BUILD -it manageiq-rpm:latest build
  popd 

  # now git clone the manageiq-pods
  git clone $MANAGEIQ_PODS_ADDR manageiq-pods --depth 1 || {
    echo "Unable to clone manageiq-pods" >&2
  }

  pushd manageiq-pods
    cp ../manageiq-rpm_build/OUTPUT/rpms/x86_64/*.rpm images/manageiq-base/rpms
    bin/build -u -l -d images -r manageiq
  popd

  git clone $MANAGEIQ_ADDR  manageiq --depth 1 || {
    echo "Unable to clone manageiq" >&2
  }
  pushd manageiq
    docker build -t manageiq/manageiq:latest .
    docker images | grep manageiq
    docker run -di -p 8443:443 manageiq/manageiq:latest 
  popd
popd
