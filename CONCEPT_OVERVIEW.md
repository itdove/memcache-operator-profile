# Runtime Code profile for Kubernetes Operators

## Abstract

The technic described in this document allows to create an instrumented operators and so deploy them in a kubernetes environment and collect the code coverage based on end-2-end tests on a composition of operators rather than on a single operator. For the simplicity of the example, we will apply this technic to only operator.

## Introduction

Often, you are asked to have a code coverage above a given percentage and you are struggling to write unit-tests. At the same time you are asked to create functional-tests and end-to-end tests but you don’t have any coverage reports on these.

Unit-tests are important but at the end of the day, we are creating operators which provide functionalities tested by the implemented  functional-tests. All these operators are packaged together to build a product on which the end-user will play a number of scenarios emulated by the end-to-end tests. These make functional tests and end-to-end tests more criticals.

We can easily generate test profile reports on unit-tests but not yet on functional-test and end-to-end tests as they are often disconnect from the operator, meaning the tests will send CRs to the operator and not call some of its methods and it is what is about in this presentation. Helping you to generate profile reports on functional-test and end-to-end tests and so increase the test code coverage percentage using the same technology. Here we will focus on the code coverage profile as other profiles such as cpu and mem are provided by the operator metrics functionality.

Tests coverage is not all as the tests themselves must check if each result of each test is accurate but it helps to focus the development of tests where it is most needed.

[Operator-sdk]:(https://sdk.operatorframework.io/docs/golang/e2e-tests) already has some capabilities but it is only to run the operator locally and with tests implemented in the same project. In this sample will also use [ginkgo](https://onsi.github.io/ginkgo/) to implement our end-2-end these tests will not have any dependencies with the operator.

### Implementation Constraints and Challenges

The are challenges to be solved:

1. "operator-sdk test" doesn't provide a way to generate instrumented binary: "go test" will be use.
2. “Go test” generates the reports only when the tests are done:  As the operator is a long running process (listener), we will have to delete the tested pod or deployment at the end of the test.
3. The report is generated in the container: A volume will be used to store the file.
We can not use `kubectl cp` because the pod will be dead by the time the `kubectl cp` runs, also some images (like ubi-minimal) don’t have tar installed and so `kubectl cp` can’t be used. If you use KiND then you will have to provide a configuration file to map a volume in KiND with the host platform.
4. The operator is embedded in an image: We will modify the dockerfile and entrypoint in order to instrument and run the operator with “Go test”.

## Process

The process is divided in 3 main phases:

1. Instrument and package the operator
2. Deploy and run the operator
3. Analyze profile

### 1. Instrument and package the operator

“Go test” offers a way to generate code profile reports based on defined tests and often so called unit-test but it can do more. In fact, it generates profile reports on all code visited during the test and we can launch a “Go test” on the main() function.

“Go test” also offers the possibility to generate instrumented code which will contain the functionality to accumulate the code profile data during the execution and at the end generates the code profile report.

#### 1.1 Add the main test method

You have to only add one file in your source code called `main_test.go` next to your current `main.go`:

[cmd/main_test.go](cmd/manager/main_test.go):
```
// Copyright (c) 2020 Red Hat, Inc.
 
// +build testrunmain
 
package main
 
import (
   "testing"
)
 
func TestRunMain(t *testing.T) {
   main()
}
```

#### 1.2 Build the instrumented binary:

Build instrumented binary: Usually the operator binary is build using `operator-sdk build $IMAGE` [README.md](README.md#buildoperator) , here we will build the binary with the `go test` and so create a new Dockerfile [Dockerfile-profile](build/Dockerfile-profile) where the standard command:
```
COPY build/_output/bin/memcached-operator ${OPERATOR}`
```
will be replaced by
```
go test -covermode=atomic -coverpkg-github.com/open-cluster-management/endpoint-operator/pkg/... -c -tags testrunmain ./cmd/manager -o build/_output/manager
```
 
The `coverpkg` parameter lists the packages for which the profile report must be done.

The `-c` requests the `go test` to create a binary instead of running the test.

The `-tags` mentions the packages that must be built for that operator.

The `-o` requests to generate a binary called `manager` as by default the generated binary name is the concatenation of the package name and `.test`.

The $IMAGE will be set with an extension `-profile` to avoid overwriting the production image.

### 2. Deploy and run the operator

#### 2.1 Entrypoint

Usually the [entrypoint](build/bin/entrypoint) is
``` 
exec ${OPERATOR} $@
```
 and a new [entrypoint-profile](build/bin/entrypoint-profile) will be created with

```
exec ${OPERATOR} -test.run “^TestRunMain$” -test.coverprofile=/tmp/profile/$HOSTNAME=`date +%s%N`.out $@
```

PS: You can add more profiles such as cpu, mem and block. Run `go help test` to see the parameters. Also check [pprof](https://github.com/google/pprof) to learn more about the available reports for these proviles.

The `test.run` specifies the test that needs to run and here “^TestRunMain$”.

The `test.coverprofile` specifies the file where the profile output must be sent. The file name is built with the time in milliseconds to make it unique and so make sure we generate a new file at each pod restart.

#### 2.2 Deployment

The [operator.yaml](deploy/operator.yaml) must be updated to customized to add the volume, securityContext... For that we will use the `kustomize` capability of `kubectl`.

An [overlays/operator.yaml](overlays/operator.yaml) will be created, which will overlay the existing [deploy/operator.yaml](deploy/operator.yaml) by:
- adding the `securityContext`
- adding the `volumes` and `volumeMounts`
- Emptying the `commands` to make sure the entrypoint will be used.

Two extra files will be added to make the customization working:
- [overlays/kustomize.yaml](overlays/kustomization.yaml)
- [deploy/kustomize.yaml](deploy/kustomization.yaml)
  
The deployment itself will be done using `kubectl apply -k overlays` instead of `kubectl apply -f deploy/operator.yaml`

In this example, we use [KiND](https://kind.sigs.k8s.io/docs/user/quick-start/) as cluster with this configuration file [build/kind-config/kind-config.yaml](build/kind-config/kind-config.yaml)

#### 2.3 Run the operator

Use the following targets to run the operator:

- `make create-cluster` to create the KiND cluster.
- `make install-profile` to install the memcached


#### 2.4 Run your tests

Once the operator is deployed you run your tests.

#### 2.5 Stop the operator and get profile

In order to get the profile we must stop the pods, here we will remove the memcached but stopping the pod will have the same effect, generate the profile file.

- `make uninstall-profile` to uninstall the memcached.
- `make delete-cluster` to delete the cluster.

The profile file will be created in the `profile` directory.

### 3 Analyze

The standard `go tools` can be used to generate the html or extract the profile percentage.

You can run `make generate-profile` for that.