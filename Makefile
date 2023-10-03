KUBEBUILDER_VERSION := 1.28.0
KUBEBUILDER_ASSETS := ~/envtest-binaries/kubebuilder/bin

.PHONY: help setup-envtest build build-binary tests kind-create-cluster

.DEFAULT_GOAL := help

help: ## Display this help message
	awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup-envtest: ## Set up the environment for testing
	# Download and setup binaries required by envtest https://book.kubebuilder.io/reference/envtest.html
	curl -sSLo envtest-bins.tar.gz "https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-$(KUBEBUILDER_VERSION)-linux-amd64.tar.gz"
	rm -rf ~/envtest-binaries
	mkdir -p ~/envtest-binaries
	tar -zvxf envtest-bins.tar.gz
	mv kubebuilder ~/envtest-binaries
	ls -ltraR ~/envtest-binaries/kubebuilder/bin
	rm -rf envtest-bins.tar.gz
	echo $$KUBEBUILDER_ASSETS

build: test build-binary build-image ## Build the application

build-binary: ## Build the binary
	go build -o controller main.go

namespace: ## Create the required namespace
	kubectl delete ns sa-controller --wait=true || true
	kubectl create ns sa-controller

run-binary: build-binary namespace ## Run the binary
	CONTROLLER_NAMESPACE=sa-controller ./controller

run: tests build-binary delete-manifest ## Run the application

test: setup-envtest ## Run the tests
	SKIP_FETCH_TOOLS=1 ACK_GINKGO_DEPRECATIONS=1.16.5 KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS) \
		go test -race -v ./... ./controllers/... -count=1 -args -ginkgo.v

kind: kind-delete-cluster kind-create-cluster build-image kind-load-image ## Manage kind cluster

kind-delete-cluster: ## Delete the kind cluster
	kind delete clusters demo || true

kind-create-cluster: ## Create a kind cluster
	kind create cluster --name demo --config manifests/kind-config.yaml --image kindest/node:v$(KUBEBUILDER_VERSION)

kind-load-image: ## Load docker image into the kind cluster
	kind load docker-image disable-automount-default-sa-controller:1.0.0  --name demo

build-image: ## Build the docker image
	docker build . --tag=disable-automount-default-sa-controller:1.0.0 --no-cache

all: tests build-image build-binary ## Run all steps

tests: clean-envtest setup-envtest test ## Clean up, set up, and run tests

clean-envtest: ## Clean up environment for testing
	rm -rfv ~/envtest-binaries/kubebuilder/bin

apply-manifest: ## Apply k8s manifests
	kubectl apply -f manifests/deployment.yaml

delete-manifest: ## Delete k8s manifests
	kubectl delete -f manifests/deployment.yaml || true

logs: ## View logs of the controller
	kubectl logs -f -n disable-automount-default-sa-controller-ns -l app=controller

deploy: delete-manifest apply-manifest ## Deploy the application
