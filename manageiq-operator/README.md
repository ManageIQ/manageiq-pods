# ManageIQ Operator

This operator manages the lifecycle of ManageIQ application on a Kubernetes or OCP4 cluster.


## Running ManageIQ under operator control.

Follow the documentation for [preparing the namespace](https://www.manageiq.org/docs/reference/latest/installing_on_kubernetes/index.html#preparing-the-kubernetes-namespace) except for the `Deploy the operator in your namespace` step.


### Run The Operator

There are three different ways the operator can be run.


+ #### Option 1: Run the latest ManageIQ Operator image from the registry in the cluster

  Follow the `Deploy the operator in your namespace` step in the documentation.

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

# Further Notes:

## Customizing the installation

See [official documentation](https://www.manageiq.org/docs/reference/latest/installing_on_kubernetes/index.html)
