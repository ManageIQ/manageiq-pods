# Deploy ManageIQ on OpenShift

[![Join the chat at https://gitter.im/ManageIQ/manageiq-pods](https://badges.gitter.im/ManageIQ/manageiq-pods.svg)](https://gitter.im/ManageIQ/manageiq-pods?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
**This guide will demo deploying ManageIQ in OpenShift as its example use-case but this method could actually be used in a different container cluster environment**

##Purpose

This example gives a base template to deploy a multi-pod ManageIQ appliance with the DB stored in a persistent volume on OpenShift. It provides a step-by-step setup including cluster administrative tasks as well as basic user information and commands. The ultimate goal of the project is to be able to decompose the ManageIQ appliance into several containers running on a pod or a series of pods.

###Prerequisites:

* OpenShift 3.x
* NFS or other compatible volume provider
* A cluster-admin user

###Cluster Sizing

In order to avoid random deployment failures due to resource starvation, we recommend a minimum cluster size for a **test** environment.

* 1 x Master node with at least 8 VCPUs and 12GB of RAM
* 2 x Nodes with at least 4 VCPUs and 8GB of RAM
* 1 x NFS server with at least 10GB of space for PV use

Other sizing considerations: 

* Recommendations assume MIQ will be the only application running on this cluster.
* Alternatively, you can provision an infrastructure node to run registry/metrics/router/logging pods.
* Each MIQ application pod will consume at least 3GB of RAM on initial deployment (blank deployment without providers).
* RAM consumption will ramp up higher depending on appliance use, once providers are added expect higher resource consumption.

###Installing

`$ git clone https://github.com/ManageIQ/manageiq-pods.git`

###Pre-deployment preparation tasks

_**As basic user**_

Login to OpenShift and create a project

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

_**Note:**_ The current MIQ image requires a privileged pod.

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

###Create persistent volumes (PVs) to host the MIQ database and application data

We need at least 3 persistent volumes to store MIQ data:

* Server   (Server specific appliance data)
* Region   (Shared region appliance data)
* Database (PostgreSQL)

Example NFS volume templates are provided (edit to match your site settings), **please skip this step you have already configured persistent storage.**

For NFS backed volumes, please ensure your NFS server firewall is configured to allow traffic on port 2049 (TCP) from the OpenShift cluster.

_**Note:**_ Recommended permissions for the PV volumes are 775, root uid/gid owned.

_**As admin:**_

```bash
$ oc create -f miq-pv-db-example.yaml
$ oc create -f miq-pv-region-example.yaml
$ oc create -f miq-pv-server-example.yaml
```
Verify PV creation
```bash
$ oc get pv
NAME       CAPACITY   ACCESSMODES   RECLAIMPOLICY   STATUS      CLAIM  REASON   AGE
miq-pv01   2Gi        RWO           Recycle         Available                   30s
miq-pv02   2Gi        RWO           Recycle         Available                   19s
miq-pv03   2Gi        RWO           Recycle         Available                   14s
```

It is strongly suggested that you validate NFS share connectivity from an OpenShift node prior attemping a deployment.

###Increase maximum number of imported images on ImageStream

By default OpenShift will import 5 images per ImageStream, we build and use more than 5 images in our repos for MIQ deployments.

You can modify these settings on the master node at /etc/origin/master/master-config.yml, add the following at the end of the file and re-start the master service:

```yaml
...
imagePolicyConfig:
  maxImagesBulkImportedPerRepository: 100
```
```bash
$ systemctl restart atomic-openshift-master
```

## Deploy MIQ

Create the MIQ template for deployment and verify it is now available in your project

_**As basic user**_

```bash
$ oc create -f templates/miq-template.yaml
template "manageiq" created
$ oc get templates
NAME       DESCRIPTION                   PARAMETERS        OBJECTS
manageiq   ManageIQ appliance template   18 (1 blank)      12

```

The supplied template provides customizable deployment parameters, use _oc process_ to see available parameters and descriptions

`$ oc process --parameters -n <your-project> manageiq`

Deploy MIQ from template using default settings

`$ oc new-app --template=manageiq`

Deploy MIQ from template using customized settings

`$ oc new-app --template=manageiq -p DATABASE_VOLUME_CAPACITY=2Gi,MEMORY_POSTGRESQL_LIMIT=4Gi`

##Verifying the setup was successful

_**Note:**_ The first deployment could take several minutes as OpenShift is downloading the necessary images.

###Confirm the MIQ pod is bound to the correct SCC

List and obtain the name of the miq-app pod

```bash
$ oc get pod
NAME                 READY     STATUS    RESTARTS   AGE
manageiq-1-fzwzm     1/1       Running   0          4m
memcached-1-6iuxu    1/1       Running   0          4m
postgresql-1-2kxc3   1/1       Running   0          4m
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
###Verify the persistent volumes are attached to postgresql and miq-app pods

```bash
$ oc volume pods --all
pods/manageiq-1-of9nl
  pvc/manageiq-server (allocated 2GiB) as miq-app-volume-server
    mounted at /persistent
  pvc/manageiq-region (allocated 2GiB) as miq-app-volume-region
    mounted at /persistent-region
  secret/default-token-nw0qi as default-token-nw0qi
    mounted at /var/run/secrets/kubernetes.io/serviceaccount
pods/postgresql-1-3fltn
  pvc/manageiq-postgresql (allocated 2GiB) as miq-pgdb-volume
    mounted at /var/lib/pgsql/data
  secret/default-token-nw0qi as default-token-nw0qi
    mounted at /var/run/secrets/kubernetes.io/serviceaccount
```

###Check readiness of the MIQ pod

_**Note:**_ Please allow ~5 minutes once pods are in Running state for MIQ to start responding on HTTPS

```bash
$ oc describe pods <miq_pod_name>
...
Conditions:
  Type		Status
  Ready 	True 
Volumes:
...
```
###Disable Image Change Triggers
By default on initial deployments the automatic image change trigger is enabled, this could potentially start an unintended upgrade on a deployment if a newer image was found in the IS.

Once you have successfully validated your MIQ deployment, disable automatic image change triggers for MIQ DCs on project.

```bash
$ oc set triggers dc --manual -l app=manageiq
deploymentconfig "manageiq" updated
deploymentconfig "memcached" updated
deploymentconfig "postgresql" updated

$ oc set triggers dc --from-config --auto -l app=manageiq
deploymentconfig "manageiq" updated
deploymentconfig "memcached" updated
deploymentconfig "postgresql" updated
```

Please note the config change trigger is kept enabled, if you desire to have full control of your deployments you can alternatively turn it off.

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

## Troubleshooting
Under normal circumstances the entire first time deployment process should take around ~10 minutes, indication of issues can be seen
by examination of the deployment events and pod logs.

### Re-trying a failed deployment

_**As basic user**_


```bash
$ oc get pods
NAME                 READY     STATUS    RESTARTS   AGE
manageiq-1-deploy    0/1       Error     0          25m
memcached-1-yasfq    1/1       Running   0          24m
postgresql-1-wfv59   1/1       Running   0          24m

$ oc deploy manageiq --retry
Retried #1
Use 'oc logs -f dc/manageiq' to track its progress.
```
Allow a few seconds for the failed pod to get re-scheduled, then begin checking events and logs

```bash
$ oc describe pods <pod-name>
...
Events:
  FirstSeen	LastSeen	Count	From							SubobjectPath			Type		Reason		Message
  ---------	--------	-----	----							-------------			--------	------		-------
15m		15m		1	{kubelet ocp-eval-node-2.e2e.bos.redhat.com}	spec.containers{manageiq}	Warning		Unhealthy	Readiness probe failed: Get http://10.1.1.5:80/: dial tcp 10.1.1.5:80: getsockopt: connection refused
```

Liveness and Readiness probe failures indicate the pod is taking longer than expected to come alive/online, check pod logs.

_**Note:**_ The miq-app container is systemd based, use _oc rsh_ instead of _oc logs_ to obtain journal dumps

`$ oc rsh <pod-name> journalctl -x`

It might also be useful to transfer all logs from the miq-app pod to a directory on the host for further examination, we can use _oc rsync_ for this task.

```bash
$ oc rsync <pod-name>:/persistent/container-deploy/log /tmp/fail-logs/
receiving incremental file list
log/
log/appliance_initialize_1477272109.log
log/restore_pv_data_1477272010.log
log/sync_pv_data_1477272010.log

sent 72 bytes  received 1881 bytes  1302.00 bytes/sec
total size is 1585  speedup is 0.81
```
