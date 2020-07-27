# ManageIQ Operator

This operator manages the lifecycle of ManageIQ application on a OCP4 cluster.

## Customizing the installation

ManageIQ can be run with an external Postgres or messaging server.  To do so, please create the required OpenShift secret(s) with the correct parameters using the template(s) for [Postgres](templates/app/postgresql-secrets.yaml) or [messaging](/templates/app/kafka-secrets.yaml) and provide those secret names as `databaseSecret` and/or `kafkaSecret` in `manageiq.org_v1alpha1_manageiq_cr.yaml`.

If you want to use a custom TLS certificate, it can be created with `oc create secret tls tls-secret --cert=tls.crt --key=tls.key` and setting the secret name as `tlsSecret` in `manageiq.org_v1alpha1_manageiq_cr.yaml`.

## Run Operator

Deploy the ManageIQ CRD

```bash
$ oc create -f deploy/crds/manageiq.org_manageiqs_crd.yaml
```

### Option A: Run Locally(outside of cluster)

In project root directory, run

```bash 
$ operator-sdk run --local --namespace=<yournamespace>
```

### Option B: Run Inside Cluster

1. Build the operator image:

```bash
$ operator-sdk build docker.io/example/manageiq-operator:latest
```

2. Update the operator deployment with the new image:

```bash
$ sed -i 's|docker.io/manageiq/manageiq-operator:latest|docker.io/example/manageiq-operator:latest|g' deploy/operator.yaml
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
$ oc create -f deploy/crds/manageiq.org_v1alpha1_manageiq_cr.yaml
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

## Creating an Operator Bundle

Create the bundle image and push to an image registry

```
$ operator-sdk bundle create docker.io/example/manageiq-bundle:0.0.1 --image-builder podman --directory deploy/olm-catalog/manageiq-operator/0.0.1/ --channels alpha --default-channel alpha
$ podman push docker.io/example/manageiq-bundle:0.0.1
```
