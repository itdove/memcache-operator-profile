.DEFAULT_GOAL:=help
SHELL:=/bin/bash
NAMESPACE=memcached

##@ Application

install-profile: ## Install all resources (CR/CRD's, RBAC and Operator)
	@echo ....... Creating namespace ....... 
	- kubectl create namespace ${NAMESPACE}
	@echo ....... Applying CRDs .......
	- kubectl apply -f deploy/crds/cache.example.com_memcacheds_crd.yaml -n ${NAMESPACE}
	@echo ....... Applying Rules and Service Account .......
	- kubectl apply -f deploy/role.yaml -n ${NAMESPACE}
	- kubectl apply -f deploy/role_binding.yaml  -n ${NAMESPACE}
	- kubectl apply -f deploy/service_account.yaml  -n ${NAMESPACE}
	@echo ....... Applying Operator .......
	- kubectl apply -k overlays --dry-run=true -o yaml | sed "s|REPLACE_IMAGE|$$IMAGE|g" | kubectl apply -n ${NAMESPACE}  -f -
	@echo ....... Creating the CRs .......
	- kubectl apply -f deploy/crds/cache.example.com_v1alpha1_memcached_cr.yaml -n ${NAMESPACE}

uninstall-profile: ## Uninstall all that all performed in the $ make install
	@echo ....... Uninstalling .......
	@echo ....... Deleting CRDs.......
	- kubectl delete -f deploy/crds/cache.example.com_memcacheds_crd.yaml -n ${NAMESPACE}
	@echo ....... Deleting Rules and Service Account .......
	- kubectl delete -f deploy/role.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/role_binding.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/service_account.yaml -n ${NAMESPACE}
	@echo ....... Deleting Operator .......
	- kubectl delete -k overlays -n ${NAMESPACE}
	@echo ....... Deleting namespace ${NAMESPACE}.......
	- kubectl delete namespace ${NAMESPACE}

##@ Development

code-vet: ## Run go vet for this project. More info: https://golang.org/cmd/vet/
	@echo go vet
	go vet $$(go list ./... )

code-fmt: ## Run go fmt for this project
	@echo go fmt
	go fmt $$(go list ./... )

code-dev: ## Run the default dev commands which are the go fmt and vet then execute the $ make code-gen
	@echo Running the common required commands for developments purposes
	- make code-fmt
	- make code-vet
	- make code-gen

code-gen: ## Run the operator-sdk commands to generated code (k8s and openapi)
	@echo Updating the deep copy files with the changes in the API
	operator-sdk generate k8s
	@echo Updating the CRD files with the OpenAPI validations
	operator-sdk generate openapi

##@ Tests

test-e2e-profile:
	$(eval MASTER_URL := $(kubectl cluster-info | grep "master" | cut -d' ' -f6))
	@ginkgo -- -master-url=$(MASTER_URL)
	
	
build-profile:
	mkdir -p build/_output/bin
	GOOS=linux GOARCH=amd64 go test -covermode=atomic -coverpkg=github.com/operator-framework/operator-sdk-samples/go/memcached-operator/pkg/... -c -tags testrunmain ./cmd/manager -o build/_output/bin/memcached-operator
	docker build . -f build/Dockerfile-profile -t $$IMAGE

create-cluster:
	rm -rf profile
	mkdir -p profile/data
	kind create cluster --name memcache-operator-cluster --config=build/kind-config/kind-config.yaml
	kind export kubeconfig --name=memcache-operator-cluster
	kind load docker-image $$IMAGE --name=memcache-operator-cluster

merge-profile:
	@gocovmerge profile/data/cover-*.out >> profile/profile.out

generate-profile:
	$(eval COVERAGE := $(shell go tool cover -func=profile/profile.out | grep "total:" | awk '{ print $$3 }' | sed 's/[][()><%]/ /g'))
	@echo "-------------------------------------------------------------------------"
	@echo "TOTAL COVERAGE IS ${COVERAGE}%"
	@echo "-------------------------------------------------------------------------"
	@go tool cover -html profile/profile.out -o profile/profile.html
	@echo "Coverage results are located at file profile/profile.html"

delete-cluster:
	kind delete cluster --name memcache-operator-cluster

demo: delete-cluster create-cluster install-profile test-e2e-profile uninstall-profile delete-cluster merge-profile generate-profile

.PHONY: help
help: ## Display this help
	@echo -e "Usage:\n  make \033[36m<target>\033[0m"
	@awk 'BEGIN {FS = ":.*##"}; \
		/^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
