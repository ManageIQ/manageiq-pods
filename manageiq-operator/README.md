# ManageIQ Operator

This operator manages the lifecycle of ManageIQ application on a Kubernetes or OCP4 cluster.

## Additional dependencies you may need for development

The following dependencies are required for some of the make commands:

- kustomize: https://github.com/kubernetes-sigs/kustomize/releases
- etcd: https://github.com/etcd-io/etcd/releases/
- kube-apiserver: `sudo dnf install kubernetes-master && sudo chmod +x /usr/bin/kube-apiserver` ???

## Running ManageIQ under operator control.

Follow the documentation for [preparing the namespace](https://www.manageiq.org/docs/reference/latest/installing_on_kubernetes/index.html#preparing-the-kubernetes-namespace) except for the `Deploy the operator in your namespace` step.

### Run The Operator

There are three different ways the operator can be run.


+ #### Option 1: Run the latest ManageIQ Operator image from the registry in the cluster

  Follow the `Deploy the operator in your namespace` step in the documentation.

+ #### Option 2: Run your own custom ManageIQ image inside the Cluster

  1 - Build and push your operator image:

    ```bash
    $ IMG=docker.io/<your_username_or_organization>/manageiq-operator make docker-build docker-push
    ```

  2 - Update the operator deployment yaml file with your custom image:

    ```bash
    $ sed -i 's|docker.io/manageiq/manageiq-operator:latest|docker.io/<your_username_or_organization>/manageiq-operator:latest|g' config/manager/manager.yaml
    ```

  3 - Run your custom image from the registry:

    ```bash
    $ oc create -f config/manager/manager.yaml
    ```

+ #### Option 3: Run locally (on your local laptop/computer, outside of the cluster)

  ```bash
  $ WATCH_NAMESPACE=<your_namespace> make run
  ```

# Further Notes:

## Customizing the installation

See [official documentation](https://www.manageiq.org/docs/reference/latest/installing_on_kubernetes/index.html)

# License

This project is available as open source under the terms of the [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0).
