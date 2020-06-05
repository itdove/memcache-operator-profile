# Runtime Code Profile for Kubernetes Operators 

The technique described in this blog allows you to create instrumented operators and deploy them in a Kubernetes environment.  You can then collect the code coverage based on end-to-end tests on a composition of operators, rather than on a single operator. For the simplicity of the example, I will apply this technique to only operators.

## Why use operators for code coverage?

Unit tests are important, but at the end of the day, we are creating operators that require _functional_ tests. All these operators are packaged together to build a product onto which the end-user tests multiple scenarios that are emulated by the end-to-end tests. These make _functional_ tests and _end-to-end_ tests more critical.

Often, you are asked to have a code coverage above a given percentage, and you are struggling to write unit-tests. At the same time you are asked to create functional tests and end-to-end tests, but you don’t have any coverage reports on these tests.

We can easily generate test profile reports on unit tests, but not yet on functional test and end-to-end tests, as they are often disconnected from the operator. This means the tests will send CRs to the operator and not call some of its methods. This problem I address in this blog by helping you to generate profile reports on functional tests and end-to-end tests to increase the test code coverage percentage using the same technology. 

Here, I will focus on the code coverage profile as other profiles such as CPU and memory are provided by the operator metrics functionality.

Tests coverage is not the final goal, as test themselves must check if each result of each test is accurate, but it helps to focus the development of tests where it is most needed.

As you can see from [Operator-sdk tests](https://sdk.operatorframework.io/docs/golang/e2e-tests), capabilities already exist, but only to run the operator locally with tests implemented in the same project. In this sample, I will also use [Ginkgo](https://onsi.github.io/ginkgo/) to implement end-to-end tests, which will have no dependencies on the operator.

## Implementation Constraints and Challenges

This technique does not come without possible constraints, which I outline in the following list:

1. The "operator-sdk test" does not provide a way to generate instrumented binary, so a "go test" is used.

2. A “Go test” generates the reports only when the tests are complete. As the operator is a long running process (listener), you need to delete the tested pod or deployment at the end of the test.

3. The report is generated in the container. A volume is used to store the file.

 Note: We cannot use `kubectl cp` because the pod is terminated by the time the `kubectl cp` runs. Also, some images, such as `ubi-minimal`, do not have `tar` installed, so `kubectl cp` cannnot be used. If you use the Kubernetes tool, _kind_, then you must provide a configuration file to map a volume in kind, with the host platform. See [Kind tool](https://kubernetes.io/docs/setup/learning-environment/kind/) for more information.

4. The operator is embedded in an image. Modify the dockerfile and entrypoint in order to instrument and run the operator with “Go test”.

## How to implement the operator

I divide the process into three main tasks, as listed:

1. First, instrument and package the operator.
2. Second, deploy and run the operator.
3. Third, analyze your profile.

### Instrument and package the operator 

A “Go test” offers a way to generate code profile reports based on defined tests and often so-called unit tests, but it can do more. In fact, it generates profile reports on all code visited during the test and you can launch a “go test” on the main function.

A “Go test” also offers the possibility to generate instrumented code, which contains the functionality to accumulate the code profile data during the execution, and at the end generate the code profile report. 

See the following procedure:

1. Add the main test method.

 You have to only add one file in your source code named, `main_test.go`, next to your current `main.go`. See [cmd/main_test.go](cmd/manager/main_test.go):

 ```
 // +build testrunmain

 package main

 import (
    "testing"
 )

 func TestRunMain(t *testing.T) {
    main()
 }
 ```

2. Build the instrumented binary.

 Usually the operator binary is built using `operator-sdk build $IMAGE`, as seen in [README.md](README.md#buildoperator). Here, I build the binary with the `go test` and so create a new [Dockerfile-profile](build/Dockerfile-profile), where the standard command is replaced.

 First  see the command to be replaced:

 ```
 COPY build/_output/bin/memcached-operator ${OPERATOR}`
 ```
 
 Replace the command with the following:

 ```
 go test -covermode=atomic -coverpkg-github.com/open-cluster-management/endpoint-operator/pkg/... -c -tags testrunmain ./cmd/manager -o build/_output/manager
 ```
 The list of definitions from the command:

  - The `coverpkg` parameter lists the packages for which the profile report must be done.

  - The `-c` requests the `go test` to create a binary instead of running the test.

  - The `-tags` mentions the packages that must be built for that operator.

  - The `-o` requests to generate a binary called `manager` as by default the generated binary name is the concatenation of the package name and `.test`.

  - The $IMAGE will be set with an extension `-profile` to avoid overwriting the production image.

### Deploy and run the operator

1. Set the Entrypoint file.

  Usually the [entrypoint](build/bin/entrypoint) resembles the following command:

  ``` 
  exec ${OPERATOR} $@
  ```
  A new [entrypoint-profile](build/bin/entrypoint-profile) is created with the following command:

  ```
  exec ${OPERATOR} -test.run “^TestRunMain$” -test.coverprofile=/tmp/profile/$HOSTNAME=`date +%s%N`.out $@
  ```

 **Tip:** You can add more profiles such as CPU, memory, and block. Run `go help test` to see the parameters. Also, check [pprof](https://github.com/google/pprof) to learn more about the available reports for these profiles.

 The `test.run` specifies the test that needs to run and here “^TestRunMain$”.

 The `test.coverprofile` specifies the file where the profile output must be sent. The file name is built with the time in milliseconds to make it unique, and so make sure we generate a new file at each pod restart.

2. Adapt the deployment YAML file.

 The [operator.yaml](deploy/operator.yaml) must be customized to add the volume, securityContext, etc..., which is used to collect the profile report. For this step, I will use the `kustomize` capability of `kubectl`.

 An [overlays/operator.yaml](overlays/operator.yaml) will be created to overwrite the existing [deploy/operator.yaml](deploy/operator.yaml) by taking the following actions:

  - Adding the `securityContext`
  - Adding the `volumes` and `volumeMounts`
  - Emptying the `commands` to make sure the entrypoint will be used

 You must add two files for customization work. Add and name the following files, `overlays` and `deploy`:

 - [overlays/kustomize.yaml](overlays/kustomization.yaml)
 - [deploy/kustomize.yaml](deploy/kustomization.yaml)

 **Important:** The deployment itself is done with the `kubectl apply -k overlays` command instead of the usual `kubectl apply -f deploy/operator.yaml` command.

 In this example, we use [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) as cluster with this configuration file [build/kind-config/kind-config.yaml](build/kind-config/kind-config.yaml).

3. Run the operator.

 Use the following targets to run the operator:

 - Run `make create-cluster` to create the kind cluster.
 - Run `make install-profile` to install the memcached.

4. Run your tests.

 Once the operator is deployed, you can run your test with the following command, which will run `ginkgo` tests:

 ```
 test-e2e-profile
 ```

5. Stop the operator and get profile

 In order to get the profile, we must stop the pods. Here, I will remove the memcached, but stopping the pod has the same effect. Generate the profile file.

 - Run `make uninstall-profile` to uninstall the memcached.
 - Run `make delete-cluster` to delete the cluster.

 The profile file is created in the `profile` directory.

### Analyze your profile

Coverage files are  available in the profile directory:

1. Run `make merge-profile`, which merges all profile files (if multiples) into one single file.

2. Run `make generate-profile`, which displays the coverage percentage and generates a `profile.html` file in the profile directory.

**Note:** The `make demo` command runs successively the following makefile targets: `delete-cluster`, `create-cluster`, `install-profile`, `test-e2e-profile`, `uninstall-profile`, `delete-cluster`, `merge-profile`, `generate-profile`.

 The standard `go tools` can be used to generate the `html` or extract the profile percentage.
 
 See [Go coverage with external tests](https://blog.cloudflare.com/go-coverage-with-external-tests/).
 
 

 Run `make generate-profile` to analyze your profile.
