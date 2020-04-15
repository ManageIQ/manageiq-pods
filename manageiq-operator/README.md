# ManageIQ Operator

This operator manages the lifecycle of ManageIQ application on a OCP4 cluster.

## Run Operator

Deploy the ManageIQ CRD

```bash
$ oc create -f deploy/crds/manageiq.org_manageiqs_crd.yaml
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

**ManageIQ Instance Example**

> The domain here will work for a Code Ready Containers cluster. Change it to one that will work for your environment.

> Additional parameters are available and documented in the Custom Resource Definition

```yaml
apiVersion: manageiq.org/v1alpha1
kind: ManageIQ
metadata:
  name: miq
spec:
  applicationDomain: "miqproject.apps-crc.testing"
```
