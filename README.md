# Deploy ManageIQ on OpenShift
**This guide will demo deploying ManageIQ in OpenShift as its example use-case but this method could actually be used in a different container cluster environment**

##Purpose

This example gives a basic template to deploy a two-pod MIQ appliance with DB stored in a persistent volume. It provides a step-by-step setup including cluster administrative tasks as well as basic user information and commands. The current MIQ image requires a **privileged** pod. The ultimate goal of the project is to be able to decompose MIQ into several containers running on a pod or a series of pods.

###Prerequisites:

* OSE 3.x
* NFS or other compatible volume provider
* A cluster-admin user

###Installing

`$ git clone https://github.com/fbladilo/miq-on-openshift.git`

###Pre-deployment preparation tasks

_**As basic user**_

Login to OSE and create a project

_**Note:**_ This section assumes you have a basic user.

`$ oc login -u <user> -p <password>`
    
   Next, create the project as follows:
   
```bash
$ oc new-project <project_name> \
--description="<description>" \
--display-name="<display_name>"
```
   
   _At a minimum, only `<project_name>` is required._

### Add your default service account to the privileged security context

The default service account for your namespace (project) must be added to the privileged SCC before they can run privileged pods.

_**As admin**_

```bash
$ oadm policy add-scc-to-user privileged system:serviceaccount:<your-namespace>:default
```

Verify that your default service account is now included in the privileged scc
```
$ oc describe scc privileged | grep Users
Users:					system:serviceaccount:openshift-infra:build-controller,system:serviceaccount:management-infra:management-admin,system:serviceaccount:management-infra:inspector-admin,system:serviceaccount:default:router,system:serviceaccount:default:registry,system:serviceaccount:<your-namespace>:default
```

###Make a persistent volume to host the MIQ database

An example NFS backed volume is provided by miq-pv-example.yaml (edit to match your settings), **please skip this step you have already configured persistent storage.**

_**As admin:**_

```bash
$ oc create -f miq-pv-example.yaml
```
Verify pv creation
```bash
$ oc get pv
NAME       CAPACITY   ACCESSMODES   STATUS      CLAIM     REASON    AGE
nfs-pv01   2Gi        RWO           Available                       24d
```
## Deploy MIQ

Create the MIQ template for deployment and verify it is now available in your project

_**As basic user**_

```bash
$ oc create -f templates/miq-template.yaml
template "manageiq" created
$ oc get templates
NAME       DESCRIPTION                   PARAMETERS        OBJECTS
manageiq   ManageIQ appliance template   8 (2 generated)   8
```

Deploy MIQ app from template

`$ oc new-app --template=manageiq`

###Confirm the Setup was Successful

_**Note:**_ The first deployment could take several minutes as OpenShift is pulling the necessary images.

###Verify the MIQ pod is bound to the correct SCC

Obtain the name of the pod

```bash
$ oc get pod
NAME                 READY     STATUS    RESTARTS   AGE
manageiq-1-3vgo0     1/1       Running   0          4m
postgresql-1-437jg   1/1       Running   0          4m
```

Export the configuration of the pod.

`$ oc export pod <miq_pod_name>`

Examine the output. Check that `openshift.io/scc` has the value `privileged`.

```yaml
...
metadata:
  annotations:
    openshift.io/scc: privileged
...
```
###Verify the persistent volume is attached to postgresql pod

```bash
$ oc volume pods <pg_pod_name>
pods/postgresql-1-437jg
  pvc/miq-pgdb-claim (allocated 2GiB) as miq-pgdb-volume
    mounted at /var/lib/pgsql/data
  secret/default-token-2se06 as default-token-2se06
    mounted at /var/run/secrets/kubernetes.io/serviceaccount
```

_**Note:**_ Please allow ~5 minutes once pods are in Running state for MIQ to start responding on HTTPS

##POD access and routes

###Get a shell on the MIQ pod

`$ oc rsh <pod_name> bash -l`

###Obtain host information from route
A route should have been deployed via template for HTTPS access on the MIQ pod

```bash
$oc get routes
NAME       HOST/PORT                       PATH      SERVICE            TERMINATION   LABELS
manageiq   miq.apps.e2e.bos.redhat.com             manageiq:443-tcp   passthrough   app=manageiq
```
Examine output and point your web browser to the reported URL/HOST.

###Note about images

The images included in this deployment were built with docker-1.9 using v1 image schema which is compatible with OSE 3.2.
Please consider this if you plan to rebuild these images with docker-1.10 or newer, the registry included in OSE 3.2 does not support them.
More details [here](https://docs.openshift.com/enterprise/3.2/release_notes/ose_3_2_release_notes.html#ose-32-asynchronous-errata-updates)
