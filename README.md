# Deploy ManageIQ on OpenShift

[![CI](https://github.com/ManageIQ/manageiq-pods/actions/workflows/ci.yaml/badge.svg?branch=oparin)](https://github.com/ManageIQ/manageiq-pods/actions/workflows/ci.yaml)
[![Join the chat at https://gitter.im/ManageIQ/manageiq-pods](https://badges.gitter.im/ManageIQ/manageiq-pods.svg)](https://gitter.im/ManageIQ/manageiq-pods?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

[![Build history for oparin branch](https://buildstats.info/github/chart/ManageIQ/manageiq-pods?branch=oparin&buildCount=50&includeBuildsFromPullRequest=false&showstats=false)](https://github.com/ManageIQ/manageiq-pods/actions?query=branch%3Amaster)

**This guide will demo deploying ManageIQ in OpenShift as its example use-case but this method could actually be used in a different container cluster environment**

## Purpose

This example gives a base template to deploy a multi-pod ManageIQ appliance with the DB stored in a persistent volume on OpenShift. It provides a step-by-step setup including cluster administrative tasks as well as basic user information and commands. The ultimate goal of the project is to be able to decompose the ManageIQ appliance into several containers.

### Prerequisites:

* Kubernetes 1.18 or higher
* A volume provider
* A cluster-admin user

### Cluster Sizing

In order to avoid random deployment failures due to resource starvation, we recommend a minimum cluster size for a **test** environment.

* 1 x Master node with at least 8 VCPUs and 12GB of RAM
* 2 x Nodes with at least 4 VCPUs and 8GB of RAM
* 20GB of storage for MIQ PV use

Other sizing considerations:

* Recommendations assume MIQ will be the only application running on this cluster.
* Alternatively, you can provision an infrastructure node to run registry/metrics/router/logging pods.

### Installing

`$ git clone https://github.com/ManageIQ/manageiq-pods.git`

### Follow the Operator deployment instructions

Production Deployment:
https://www.manageiq.org/docs/reference/latest/installing_on_kubernetes/index.html#preparing-the-kubernetes-namespace

Development Deployment:
[README](manageiq-operator/README.md)


## Building Images
The bin/build script will build the entire chain of images.

The script requires at a minimum the `-d` option to specify the location of the `images` directory, (`./images` if run from the repo root) and the `-r` option to specify the resulting image repo and namespace.

For example, if you wanted to build all the images tagged as `manageiq/<image-name>:latest`, you would run the following command from the repo root.

```bash
./bin/build -d . -r manageiq
```

Additional options are also available:
  - `-n` Use the --no-cache option when running the manageiq-base image build
  - `-p` Push the images after building
  - `-s` Run a release build, using the latest tagged rpm build, excluding nightly rpm builds
  - `-t <tag>` Tag the built images with the specified tag (default: latest)

### Building RPMs locally

If you want to build images containing your fork or different branches of ManageIQ source code, the `-b` option can be used, which will build RPMs locally before building the images. RPMs are built in `manageiq/rpm_build` container image but a different image can be used if needed.

For example, if you want to build RPMs using `test/rpm_build` image and override rpm_build options using `my_options/options.yml`, you would run:

```bash
RPM_BUILD_IMAGE=test/rpm_build RPM_BUILD_OPTIONS=my_options bin/build -d . -r manageiq -b
```

An example options.yml to use "feature" branch on the "example" fork of ManageIQ/manageiq:

```
repos:
  manageiq:
    url:  https://github.com/example/manageiq.git
    ref:  feature  (<-- must be tag or branch)
```

Note: `RPM_BUILD_OPTIONS` is a relative path to the directory where options.yml is.

Refer to https://github.com/ManageIQ/manageiq-rpm_build for rpm build details.

### Using locally built RPMs

If you already have locally built RPMs and want to use them instead of using RPMs from manageiq yum repo, copy `<arch>` directory to images/manageiq-base/rpms:

`images/manageiq-base/rpms/x86_64/manageiq-*.rpm`

Then, run:

```bash
bin/build -d . -r manageiq -l
```

### Including the VMware VDDK in the image build

Download the VMware VDDK library and move or copy it into the `images/manageiq-base-worker/container-assets/` directory before starting the image build.  If found, the image build will copy the files into the desired location and link the shared objects for you.
