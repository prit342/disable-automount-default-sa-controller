# disable-automount-default-sa-controller

- The repo houses a kubernetes controller that watches the `default` service account across all namespaces and sets the `automountServiceAccount` field to false
- By setting `automountServiceAccountToken` to `false` for all default service accounts, the controller fulfills the control 5.1.5 set by
[CIS Kubernetes benchmark](https://www.cisecurity.org/benchmark/kubernetes) 
- The controller is based on the example controllers available [here](https://github.com/kubernetes-sigs/controller-runtime/tree/master/examples)

## Prerequisites

- You will need to install [`kind`](https://kind.sigs.k8s.io/docs/user/quick-start/)
- You will also need to install `curl`, `docker`, `make` and `kubectl`

## Running tests

- Test uses the env test binaries and can be run locally using the following make target:

```bash
make tests
```

## Deploying & testing the controller in a local Kind cluster

- You can build and run the controller in a local kind cluster using the following make target:

```bash
  make kind
```

- The above command will create a new Kind cluster called `demo` based on kubernetes version `1.33.1` and will build and import the Docker image into the Kind nodes

- Once the docker image is loaded into the Kind cluster, you can run it as a Kubernetes deployment using the following make target:

```bash
  make deploy
```

- Check the logs from the controller using the following command:

```bash
  make logs
```

- To test the controller, you can create a new namespace and check the default service account in that namespace:

```bash
kubectl create namespace test-namespace

kubectl get serviceaccount default -n test-namespace -o yaml
```

*You should see the `automountServiceAccountToken` field set to `false` in the output of the above command*

Output:
```
apiVersion: v1
automountServiceAccountToken: false
kind: ServiceAccount
metadata:
  creationTimestamp: "2025-07-01T16:36:52Z"
  name: default
  namespace: test-namespace
  resourceVersion: "2450"
  uid: a25876f2-6ecd-4d9a-ac48-c6ebc0ea49bb
```

- If you patch the service account to set `automountServiceAccountToken` to `true`, the controller will automatically revert it back to `false`:

```bash
kubectl patch serviceaccount default -n test-namespace --type='json' -p='[{"op": "replace", "path": "/automountServiceAccountToken", "value": true}]'
```


- Cleanup the test cluster

```bash
make kind-delete-cluster
```
