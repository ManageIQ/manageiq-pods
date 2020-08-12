# ManageIQ Operator

This operator manages the lifecycle of ManageIQ application on a OCP4 cluster.


## Running ManageIQ under operator control.

There are five high level steps for running ManageIQ under operator control:

  + Step 1. Deploy the ManageIQ Custom Resource Definition (CRD) (If it is not already.)
  + Step 2. Set up RBAC
  + Step 3. Run The Operator
  + Step 4. Perform any optional custom configurations
  + Step 5. Run ManageIQ by creating the Custom Resource (CR)

The details of the five steps are as follows:

### Step 1. Deploy the ManageIQ Custom Resource Definition (CRD)

The ManageIQ CRD needs to be defined on the cluster.  If it is already available it does not need to be created again.

To determine if it is already available execute command:

```bash
$ oc get crds | grep manageiqs.manageiq.org
```

If it is not already available it can be deployed with command:

```bash
$ oc create -f deploy/crds/manageiq.org_manageiqs_crd.yaml
```

### Step 2. Set up RBAC

```bash 
$ oc create -f deploy/role.yaml
$ oc create -f deploy/role_binding.yaml
$ oc create -f deploy/service_account.yaml
```

### Step 3. Run The Operator

There are three different ways the operator can be run.


+ #### Option 1: Run the latest ManageIQ Operator image from the registry in the cluster

  The default in the operator.yaml is for the latest manageiq-operator image.
  So no change is required. Simply create the operator.

  ```bash
  $ oc create -f deploy/operator.yaml
  ```

+ #### Option 2: Run your own custom ManageIQ image inside the Cluster

  1 - Build your operator image:

    ```bash
    $ operator-sdk build docker.io/<your_username_or_organization>/manageiq-operator:latest
    ```

  2 - Push your new custom image to the registry:

    ```bash
    $ docker push docker.io/<your_username_or_organization>/manageiq-operator:latest
    ```
    
  3 - Update the operator deployment yaml file with your custom image:

    ```bash
    $ sed -i 's|docker.io/manageiq/manageiq-operator:latest|docker.io/<your_username_or_organization>/manageiq-operator:latest|g' deploy/operator.yaml
    ```

  4 - Run your custom image from the registry:

    ```bash
    $ oc create -f deploy/operator.yaml
    ```

+ #### Option 3: Run locally (on your local laptop/computer, outside of the cluster)

  ```bash
  $ operator-sdk run --local --namespace=<your namespace>
  ```

### Step 4. Perform any optional custom configurations

*see the OpenID-Connect example below*

### Step 5. Run ManageIQ by creating the Custom Resource (CR)

```bash
$ oc create -f deploy/crds/manageiq.org_v1alpha1_manageiq_cr.yaml
```

# Further Notes:

## Customizing the installation

ManageIQ can be run with an external Postgres or messaging server.  To do this, create the required OpenShift secret(s) with the correct parameters using the template(s) for [Postgres](templates/app/postgresql-secrets.yaml) or [messaging](/templates/app/kafka-secrets.yaml) and provide those secret names as `databaseSecret` and/or `kafkaSecret` in `manageiq.org_v1alpha1_manageiq_cr.yaml`.

If you want to use a custom TLS certificate, one can be created with:

```bash
oc create secret tls tls-secret --cert=tls.crt --key=tls.key` and setting the secret name as `tlsSecret` in `manageiq.org_v1alpha1_manageiq_cr.yaml`.
```

## Creating an Operator Bundle

Create the bundle image and push to an image registry

```
$ operator-sdk bundle create docker.io/example/manageiq-bundle:0.0.1 --image-builder podman --directory deploy/olm-catalog/manageiq-operator/0.0.1/ --channels alpha --default-channel alpha
$ podman push docker.io/example/manageiq-bundle:0.0.1
```
## Configuring the application domain name

Modify `deploy/crds/manageiq.org_v1alpha1_manageiq_cr.yaml` as follows:

**Note:** The domain here will work for a Code Ready Containers cluster. Change it to one that will work for your environment.
Additional parameters are available and documented in the Custom Resource Definition

```yaml
apiVersion: manageiq.org/v1alpha1
kind: ManageIQ
metadata:
  name: miq
spec:
  applicationDomain: "miqproject.apps-crc.testing"
```

## Configuring OpenID-Connect Authentication

To run ManageIQ with OpenID-Connect Authentication, include these steps at **Step 4. Perform any optional custom configurations** from above.

For this example we tested with Keycloak version 11.0

+ Create a secret containing the OpenID-Connect's `Client ID` and `Client Secret`

You pick the name for `<name of your openshift client secret>`
The values for `CLIENT_ID` and `CLIENT_SECRET` come from your Keycloak client definition.

```bash
$ oc create secret generic <name of your openshift client secret>  \
  --from-literal=CLIENT_ID=<your Keycloak client ID>  \
  --from-literal=CLIENT_SECRET=<your Keycloak client secret>
```

+ Modify the Custom Resource (CR) .yaml file to identify your OpenID-Connect provider URL and the secret just created.

Here is an example of `deploy/crds/manageiq.org_v1alpha1_manageiq_cr.yaml` for OpenID-Connect authentication:

```yaml
apiVersion: manageiq.org/v1alpha1
kind: ManageIQ
metadata:
  name: miq
spec:
  applicationDomain: "miqproject.apps-crc.testing"
  httpdAuthenticationType: openid-connect
  oidcProviderURL: https://<your keycloak FQDN>/auth/realms/<your Keycloak Realm>/.well-known/openid-configuration
  oidcClientSecret: <name of your openshift client secret> 
```

