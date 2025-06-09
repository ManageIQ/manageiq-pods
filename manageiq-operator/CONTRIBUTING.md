# Contributing

Please read the [ManageIQ guides](http://github.com/ManageIQ/guides) before contributing.

## Development Setup

### Pre-requisites

* In order to compile the operator, install the operator-sdk:

   |      |     |
   | ---- | --- |
   | dnf  | [Installation Docs](https://sdk.operatorframework.io/docs/installation/#install-from-github-release) |
   | brew | `brew install operator-sdk` |

* Ensure you have a Kubernetes or OpenShift cluster available in which you can test. Set up of the cluster is outside the scope of this document.

#### Cluster pre-requisites

* Create a namespace for deployment and development.
* Ensure you have any ImagePullSecrets configured.
* Ensure you have Physical Volumes created and available for the various services, such as PostgreSQL.
* Install the [Strimzi](https://strimzi.io/) operator into the cluster to be used as the Kafka service.

#### ManageIQ operator setup

* If you want to develop and run the operator locally, acting on the remote cluster namespace, be sure that the requisite Cluster Resource Definition (CRD), Role, Service Account, and Role Binding are installed in the cluster.

  ```
  kubectl apply -f config/crd/bases/manageiq.org_manageiqs.yaml
  kubectl apply -f config/rbac/role.yaml
  kubectl apply -f config/rbac/service_account.yaml
  kubectl apply -f config/rbac/role_binding.yaml
  ```

## Development

To run the operator locally in development:

```sh
$ WATCH_NAMESPACE=<namespace> make run
```

If you don't already have a ManageIQ CR in the namespace, now you can create one, and the operator will see that and start acting accordingly.
