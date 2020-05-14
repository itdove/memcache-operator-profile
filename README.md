# Memcached Go Operator Runtime Code Coverage Sample

## Overview

This Memcached operator code coverage is an example on how to generate code coverage or profile report on an operator, the [memcache-operator](https://github.com/operator-framework/operator-sdk-samples/tree/master/go/memcached-operator) was taken as base for this example.

## Prerequisites

- [go][go_tool] version v1.13+.
- [docker][docker_tool] version 17.03+
- [kubectl][kubectl_tool] v1.14.1+
- [operator-sdk][operator_install]
- Access to a Kubernetes v1.14.1+ cluster or use [KiND](https://kind.sigs.k8s.io/docs/user/quick-start/)

## Overview of the code profile for a Kubernetes Operator

[Runtime Code profile for Kubernetes Operators](CONCEPT_OVERVIEW.md)

## Quick Demo


1. `go mod tidy`
2. `export IMAGE=<your image name>` (ie: "quay.io/example-inc/memcached-operator-profile:v0.0.1")
3. `make build-profile` to build the instrumented operator.
4. `make demo` to run the entired demo which will run the following target  
   1. `make create-cluster` to create a KiND cluster
   2. `make install-profile` to install memcached instrumented operator
   3. `make uninstall-profile` to uninstall memcached instrumented operator.
   4. `make delete-cluster` to delete the KiND cluster.
   5. `make generate-profile` to generate the profile htlm and get the profile percentage.

## Getting Started

### Pulling the dependencies

Run the following command

```
$ go mod tidy
```

<a name="build-operator"></a>

### Building the operator

Build the Memcached operator image and push it to a public registry, such as quay.io:

```
$ export IMAGE=quay.io/example-inc/memcached-operator:v0.0.1
$ operator-sdk build $IMAGE
$ docker push $IMAGE
```

### Building the operator with profile

Build the Memcached operator image and push it to a public registry, such as quay.io:

```
$ export IMAGE=quay.io/example-inc/memcached-operator-profile:v0.0.1
$ make build-profile
$ docker push $IMAGE # Optional, only if you use Kube cluster instead of KiND.
```

### Using the image

```
# Update the operator manifest to use the built image name (if you are performing these steps on OSX, see note below)
$ sed -i 's|REPLACE_IMAGE|quay.io/example-inc/memcached-operator-profile:v0.0.1|g' deploy/operator.yaml
# On OSX use:
$ sed -i "" 's|REPLACE_IMAGE|quay.io/example-inc/memcached-operator-profile:v0.0.1|g' deploy/operator.yaml
```

**NOTE** The `quay.io/example-inc/memcached-operator-profile:v0.0.1` is an example. You should build and push the image for your repository.

### Deploy a KiND cluster
```shell
rm -rf /tmp/profile
mkdir -p /tmp/profile
kind create cluster --name memcache-operator-cluster --config=build/kind-config/kind-config.yaml
kind export kubeconfig --name=memcache-operator-cluster
kind load docker-image quay.io/example-inc/memcached-operator-profile:v0.0.1 --name=memcache-operator-cluster
```

### Installing

Run `make install-profile` to install the operator. Check that the operator is running in the cluster, also check that the example Memcached service was deployed.

Following the expected result.

```shell
$ kubectl get all -n memcached
NAME                                      READY   STATUS    RESTARTS   AGE
pod/example-memcached-7c4df9b7b4-lzd6j    1/1     Running   0          64s
pod/example-memcached-7c4df9b7b4-wbtkz    1/1     Running   0          64s
pod/example-memcached-7c4df9b7b4-wt6jb    1/1     Running   0          64s
pod/memcached-operator-56f54d84bf-zrtfv   1/1     Running   0          69s

NAME                                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/example-memcached            ClusterIP   10.108.124.47   <none>        11211/TCP           63s
service/memcached-operator-metrics   ClusterIP   10.108.67.82    <none>        8383/TCP,8686/TCP   66s

NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/example-memcached    3/3     3            3           65s
deployment.apps/memcached-operator   1/1     1            1           70s

NAME                                            DESIRED   CURRENT   READY   AGE
replicaset.apps/example-memcached-7c4df9b7b4    3         3         3       65s
replicaset.apps/memcached-operator-56f54d84bf   1         1         1       70s
```

### Uninstalling

To uninstall all that was performed in the above step run `make uninstall-profile`.

### Troubleshooting

Use the following command to check the operator logs.

```shell
kubectl logs deployment.apps/memcached-operator -n memcached
```

### Running Tests

Run `make test-e2e` to run the integration e2e tests with different options. For
more information see the [writing e2e tests][golang-e2e-tests] guide.

[dep_tool]: https://golang.github.io/dep/docs/installation.html
[go_tool]: https://golang.org/dl/
[kubectl_tool]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[docker_tool]: https://docs.docker.com/install/
[operator_sdk]: https://github.com/operator-framework/operator-sdk
[operator_install]: https://sdk.operatorframework.io/docs/install-operator-sdk/
[golang-e2e-tests]: https://sdk.operatorframework.io/docs/golang/e2e-tests/
