# Upgrading

This document will outline the steps to upgrade between stable releases of ManageIQ in OpenShift

## Upgrading to Hammer

This upgrade will utilize the built-in capabilities of the `oc replace` command to replace existing objects with their updated versions and deploy new images. The high level steps look like this:

In the ManageIQ Hammer release the Deployment Config, Service, and Secrets for the Embedded Ansible service are managed entirely by the application. To ensure this works properly, the Embedded Ansible role should be de-activated before the upgrade is started.

### Retrieve Secrets

During template resolution, secrets such as the admin password, encryption key, and database password will be re-generated if they are not provided. Because of this we need to ensure we bring forward the same values which were set up in the previous version.

To retrieve the secret object run `oc export secret manageiq-secrets > my_secrets.yml`. We will overlay these values on top of the generated ones after the upgrade.

### Patch OpenShift Objects

In this step, we will use the new template to patch the existing OpenShift object specs.

- Scale the application stateful sets to 0 replicas
  - `oc scale statefulset manageiq --replicas=0`
  - `oc scale statefulset manageiq-backend --replicas=0`
- Apply the changes to the project
  - Note: If any parameter modifications were made in the original call to `oc new-app` they must be set to the same values in the `oc process` command here
  - `oc process -p APPLICATION_REPLICA_COUNT=0 -l app=manageiq,template=manageiq -f templates/miq-template.yaml | oc replace -f -`
- Replace the secret using the file from the "Retrieve Secrets" step
  - `oc replace -f my_secrets.yml`
- Redeploy the PG pod to ensure the password from the old secret is used
  - `oc rollout latest postgresql`
- Delete the ansible objects
  - `oc delete deploymentconfig ansible`
  - `oc delete service ansible`
- Scale the application stateful sets back to their previous values
  - The first "manageiq" replica will migrate the database, then the rest of the replicas will come up
- Remove the old ansible secret after the migration has finished
  - `oc delete secret ansible-secrets`
