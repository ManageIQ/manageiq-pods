# ManageIQ Operator

This operator manages the lifecycle of ManageIQ application on a OCP4 cluster.

## Run Operator

Deploy the ManageIQ CRD

```bash
$ oc create -f deploy/crds/manageiq_v1alpha1_manageiq_crd.yaml
```

### Option A: Run Locally(outside of cluster)

In project root directory, run

```bash 
$ operator-sdk up local --namespace=<yournamespace>
```

### Option B: Run Inside Cluster

1. Build the operator image:

```bash
$ operator-sdk build docker.io/example/manageiq-operator:latest
```

2. Update the operator deployment with the new image:

```bash
$ sed -i 's|docker.io/manageiq/manageiq-operator:v0.0.1|docker.io/example/manageiq-operator:latest|g' deploy/operator.yaml
```

3. Push the new image to the registry:

```bash
$ docker push docker.io/example/manageiq-operator:latest
```

### Setup RBAC and deploy the operator

```bash 
$ oc create -f deploy/role.yaml
$ oc create -f deploy/role_binding.yaml
$ oc create -f deploy/service_account.yaml
$ oc create -f deploy/operator.yaml
```

### Create the CR to deploy ManageIQ

```bash
$ oc create -f deploy/crds/manageiq_v1alpha1_manageiq_cr.yaml
```

**Manageiq Instance Example**

> Deployments' resource requests here are tailered to make them fit into a crc cluster, change them according to your cluster's resource capacity*

```yaml
apiVersion: manageiq.example.com/v1alpha1
kind: Manageiq
metadata:
  name: miq
spec:
  # Add fields here
  appName:  "manageiq"
  applicationAdminPassword: "smartvm" 
  applicationDomain: "miqproject.apps-crc.testing"

  databaseName: "vmdb_production"
  databasePort: "5432"
  databaseUser: "root"
  databasePassword: "redhat"
  databaseRegion: "0"
  databaseVolumeCapacity: 15Gi

  httpdCPUReq: 100m
  httpdImageName: manageiq/httpd
  httpdImageTag: latest
  httpdMemLimit: 200Mi 
  httpdMemReq: 100Mi

  memcachedCPUReq: 200m
  memcachedImageName: manageiq/memcached
  memcachedImageTag: latest 
  memcachedMaxConnection: "1024"
  memcachedMaxMemory: "64"
  memcachedMemLimit: 256Mi
  memcachedMemReq: 64Mi
  memcachedSlabPageSize: 1m

  orchestratorCPUReq: 100m
  orchestratorImageName: manageiq-orchestrator
  orchestratorImageNamespace: manageiq
  orchestratorImageTag: latest
  orchestratorMemLimit: 16Gi
  orchestratorMemReq: 150Mi

  postgresqlCPUReq: 100m
  postgresqlImgName: docker.io/manageiq/postgresql
  postgresqlImgTag: latest
  postgresqlMaxConnections: "1000" 
  postgresqlMemLimit: 8Gi
  postgresqlMemReq: 200Mi
  postgresqlSharedBuffers: 1GB
```
