# Deploy ManageIQ on OpenShift

[![CI](https://github.com/ManageIQ/manageiq-pods/actions/workflows/ci.yaml/badge.svg)](https://github.com/ManageIQ/manageiq-pods/actions/workflows/ci.yaml)
[![Join the chat at https://gitter.im/ManageIQ/manageiq-pods](https://badges.gitter.im/ManageIQ/manageiq-pods.svg)](https://gitter.im/ManageIQ/manageiq-pods?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

[![Build history for master branch](https://buildstats.info/github/chart/ManageIQ/manageiq-pods?branch=master&buildCount=50&includeBuildsFromPullRequest=false&showstats=false)](https://github.com/ManageIQ/manageiq-pods/actions?query=branch%3Amaster)

**This guide will demo deploying ManageIQ in OpenShift as its example use-case but this method could actually be used in a different container cluster environment**

## Purpose

This example gives a base template to deploy a multi-pod ManageIQ appliance with the DB stored in a persistent volume on OpenShift. It provides a step-by-step setup including cluster administrative tasks as well as basic user information and commands. The ultimate goal of the project is to be able to decompose the ManageIQ appliance into several containers.

### Prerequisites:

* OpenShift Origin 1.5 or higher
* NFS or other compatible volume provider
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

### Pre-deployment preparation tasks

_**As basic user**_

Login to OpenShift and create a project

_**Note:**_ This section assumes you have a basic user.

`$ oc login -u <user> -p <password>`

   Next, create the project as follows:

```bash
$ oc new-project <project_name>
```

### Follow the Operator deployment instructions
[README](manageiq-operator/README.md)

### Check readiness of the MIQ pods

_**Note:**_ Please allow ~5 minutes once pods are in Running state for MIQ to start responding on HTTPS

The READY column denotes the number of replicas and their readiness state

```bash
$ oc get pods
NAME                                     READY     STATUS    RESTARTS   AGE
httpd-754985464b-4dzzx                   1/1       Running   0          37s
manageiq-orchestrator-5997776478-vx4v9   1/1       Running   0          37s
memcached-696479b955-67fs6               1/1       Running   0          37s
postgresql-5f954fdbd5-tnlmf              1/1       Running   0          37s
```

Once the database has been migrated and the orchestrator pod is up and running, it will begin to start worker pods.
After a few minutes you can see the initial set of worker pods has been deployed and the user interface should be accessible.

```bash
$ oc get pods
NAME                                     READY     STATUS    RESTARTS   AGE
event-handler-747574c54c-xpcvf           1/1       Running   0          32m
generic-55cc84f79d-gwf5v                 1/1       Running   0          32m
generic-55cc84f79d-w4vzs                 1/1       Running   0          32m
httpd-754985464b-4dzzx                   1/1       Running   0          37m
manageiq-orchestrator-5997776478-vx4v9   1/1       Running   0          37m
memcached-696479b955-67fs6               1/1       Running   0          37m
postgresql-5f954fdbd5-tnlmf              1/1       Running   0          37m
priority-7b6666cdcd-5hkkm                1/1       Running   0          32m
priority-7b6666cdcd-rcf7l                1/1       Running   0          32m
remote-console-6958c4cc7b-5kmmj          1/1       Running   0          32m
reporting-85c8488848-p5fb6               1/1       Running   0          32m
reporting-85c8488848-z7kjp               1/1       Running   0          32m
schedule-6fd7bc5688-ptsxp                1/1       Running   0          32m
ui-5b8c86f6f9-jhd9w                      1/1       Running   0          32m
web-service-858f55f55d-5tmcr             1/1       Running   0          32m

```

## Scale MIQ

ManageIQ worker deployments can be scaled from within the application web console.
Navigate to Configuration -> Server -> Workers tab to change the number of worker replicas.

Additional workers for provider operations will be deployed or removed by the orchestrator as providers are added or removed and as roles change.

_**Note:**_ The orchestrator will enforce its desired state over the worker replicas. This means that any changes made to desired replica numbers in the OpenShift UI will be quickly reverted by the orchestrator. 

## Pod access and ingress

### Get a shell on the MIQ pod

`$ oc rsh <pod_name> bash -l`

### Obtain host information from route
An ingress should have been deployed via template for HTTPS access on the MIQ pod
When an ingress is deployed in OpenShift, a route is automatically created.

```bash
$ oc get ingress
NAME      HOSTS                              ADDRESS   PORTS     AGE
httpd     miq-dev.apps.example.com                     80, 443   56s
$ oc get routes
NAME          HOST/PORT                          PATH      SERVICES   PORT      TERMINATION     WILDCARD
httpd-qlvmj   miq-dev.apps.example.com           /         httpd      80        edge/Redirect   None
```
Examine output and point your web browser to the reported URL/HOST.

### Logging In

Per the ManageIQ project [basic configuration](http://manageiq.org/docs/get-started/basic-configuration) documentation, you can now login to the MIQ web interface using the default username and password (`admin`/`smartvm`).

## Troubleshooting
Under normal circumstances the entire first time deployment process should take around ~10 minutes, indication of issues can be seen
by examination of the deployment events and pod logs.


### Allow docker.io/manageiq images in kubernetes

Depending on your cluster's configuration, kubernetes may not allow deployment of images from `docker.io/manageiq`.  If so, deploying the operator may raise an error:

```
Error from server: error when creating "deploy/operator.yaml": admission webhook "trust.hooks.securityenforcement.admission.xxx" denied the request:
Deny "docker.io/manageiq/manageiq-operator:latest", no matching repositories in ClusterImagePolicy and no ImagePolicies in the "YYY" namespace
```

To allow images from `docker.io/manageiq`, edit the clusterimagepolicies and add `docker.io/manageiq/*` to the list of allowed repositories:

```
kubectl edit clusterimagepolicies $(kubectl get clusterimagepolicies --no-headers | awk ‘{print $1}’)
```

For example:

```
...
spec:
  repositories:
  ...
  - name: docker.io/icpdashdb/*
  - name: docker.io/istio/proxyv2:*
  - name: docker.io/library/busybox:*
  - name: docker.io/manageiq/*
  ...
```

After saving this change, `docker.io/manageiq` image deployment should now be allowed.

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

Additionally the source fork and git ref for manageiq-appliance-build can be set using the following environment variables:
  - `BUILD_REF`
  - `BUILD_ORG`

A more complicated example would be to build and push all the images to the quay.io repository "test" using the source from the "feature" branch on the "example" fork:

```bash
BUILD_ORG=example BUILD_REF=feature ./bin/build -d . -r quay.io/test -p
```

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
