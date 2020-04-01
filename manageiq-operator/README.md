# ManageIQ Operator

This operator manages the lifecycle of ManageIQ application on a OCP4 cluster.

## Run Operator

### Option A: Run Locally(outside of cluster)

In project root directory, run

```bash 
$ operator-sdk up local --namespace=<yournamespace>
```

### Option B: Run Inside Cluster

Please refer to *"4. Build and run the Operator"* in [this guide](https://docs.openshift.com/container-platform/4.1/applications/operator_sdk/osdk-getting-started.html) for remaining steps


#### Deploy CRD resources

```bash
$ oc create -f deploy/crds/manageiq_v1alpha1_manageiq_crd.yaml
```

#### Create ManageIQ Custom Resource

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
