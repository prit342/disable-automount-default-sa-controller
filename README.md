# disable-automount-default-sa-controller

- The repo houses a kubernetes controller that watches the `default` service account across all namespaces and sets the `automountServiceAccount` field to false
- By setting `automountServiceAccountToken` to `false` for all default service accounts, the controller fulfills the control 5.1.5 set by
[CIS Kubernetes benchmark](https://www.cisecurity.org/benchmark/kubernetes) 
- The controller is based on the example controllers available [here](https://github.com/kubernetes-sigs/controller-runtime/tree/master/examples)

## Prerequisites

- You will need to install [`kind`](https://kind.sigs.k8s.io/docs/user/quick-start/) and its prerequisites for local testing
- You will also need to install `curl`, `docker`, `make` and `kubectl`

## Running tests

- Test uses the env test binaries and can be run locally using the following make target:

```bash
make tests
```

## Deploying the controller in a local Kind cluster

- You can build and run the controller in the local kind cluster using the following commands:

```bash
  make kind
```

- The above command will create a new Kind cluster called `demo` based on kubernetes version `1.25.0` and will build and import the Docker image into the Kind nodes

- Once the docker image is loaded into the Kind cluster, you can run it as a Kubernetes deployment using the following make target:

```bash
  make deploy
```

- Check the logs from the controller using the following command:

```bash
  make logs
```

- Cleanup the test cluster

```bash
make kind-delete-cluster
```
