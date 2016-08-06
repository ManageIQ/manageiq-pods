# Deploy ManageIQ on OpenShift
**This guide will demo deploying ManageIQ in OpenShift as its example use-case but this method could actually be used in a different container cluster environment**

##Purpose

This example gives a basic template to deploy a single pod MIQ appliance with DB stored in a persistent volume. It provides a step-by-step setup including cluster administrative tasks as well as basic user information and commands. The current MIQ image requires a **privileged** pod. The ultimate goal of the project is to be able to decompose MIQ into several containers running on a pod or a series of pods.

###Assumptions:

* OSE 3.x
* NFS or other compatible volume provider
* A cluster-admin user

##Login as basic-user and create user project

_**Note:**_ This section assumes you have a basic user.

`$ oc login -u <user> -p <password>`
    
   Next, create the project as <user>:
   
```bash
$ oc new-project <project_name> \
--description="<description>" \
--display-name="<display_name>"
```
   
   _At a minimum, only `<project_name>` is required._

## Add your default service account to privileged security context

The default service account for your namespace (project) must be added to the privileged SCC before they can run privileged pods.

_**As admin**_

```bash
$ oadm policy add-scc-to-user privileged system:serviceaccount:<your-namespace>:default
```

Verify that your default service account is now included in the privileged scc
```
$ oc describe scc privileged | grep Users
Users:					system:serviceaccount:openshift-infra:build-controller,system:serviceaccount:management-infra:management-admin,system:serviceaccount:management-infra:inspector-admin,system:serviceaccount:default:router,system:serviceaccount:default:registry,system:serviceaccount:miq:default
```

##Make a persistent volume to host the MIQ database

An example NFS backed volume is provided by miq-pv.yaml (edit to match your settings), **please skip this step you have already configured persistent storage**

_**As admin:**_

```bash
$ oc create -f miq-pv.yaml
```
Verify pv creation
```
$ oc get pv
```
## Deploy MIQ

Create the MIQ template for deployment use

`$ oc create -f templates/miq-template-monolithic.yaml`

Deploy MIQ pod from template

`$ oc new-app --template=manageiq-monolithic.yaml`

##Confirm the Setup was Successful

###Verify the Pod is Bound to the Correct scc

Get the pod name

```bash
$ oc get pod
NAME               READY     STATUS    RESTARTS   AGE
manageiq-1-pxhc5   1/1       Running   0          5h
```

Export the configuration of the pod.

`$ oc export pod <pod_name>`

Examine the output. Check that `openshift.io/scc` has the value `privileged`.

```yaml
...
metadata:
  annotations:
    openshift.io/scc: privileged
...
```
Please allow a few minutes for the database to be prepared and MIQ start responding

##POD access and routes

###Get a shell on MIQ pod

`$ oc rsh <pod_name>`

###Obtain host information from route
A route should have been deployed via template for HTTPS access on the MIQ pod

```bash
$oc get routes
NAME       HOST/PORT                       PATH      SERVICE            TERMINATION   LABELS
manageiq   miq.apps.e2e.bos.redhat.com             manageiq:443-tcp   passthrough   app=manageiq
```
Examined output and use the supplied HOST information and point your web browser to the URL reported
