
KUBEBUILDER_VERSION := 1.33.0
KIND_CLUSTER_NAME := demo
KIND_NODE_VERSION := 1.33.1

# Point to local cache of envtest binaries, populated by setup-envtest tool
KUBEBUILDER_ASSETS ?= $(shell echo $$HOME/.local/share/kubebuilder-envtest/k8s/$(KUBEBUILDER_VERSION)/bin)

.PHONY: help build build-binary test setup-envtest clean-envtest kind-create-cluster

.DEFAULT_GOAL := help

help: ## Display this help message
	awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup-envtest: ## Download envtest binaries and export KUBEBUILDER_ASSETS
	# Install the setup-envtest tool (if not already installed).
	go install "sigs.k8s.io/controller-runtime/tools/setup-envtest@latest"
	# Use setup-envtest to download the Kubernetes control plane binaries (etcd, kube-apiserver, etc) for the specified version.
	@eval "$$(setup-envtest use -p env $(KUBEBUILDER_VERSION))" && \
	echo "KUBEBUILDER_ASSETS set to $$KUBEBUILDER_ASSETS"

build: test build-binary build-image ## Build the application

build-binary: ## Build the binary
	go build -o controller main.go

run-binary: build-binary namespace ## Run the binary
	CONTROLLER_NAMESPACE=sa-controller ./controller

test: setup-envtest ## Run tests with KUBEBUILDER_ASSETS set
	@eval "$$(setup-envtest use -p env $(KUBEBUILDER_VERSION))" && \
	KUBEBUILDER_ASSETS=$$KUBEBUILDER_ASSETS go test -race -v ./controllers/... -count=1 -args -ginkgo.v

kind: kind-delete-cluster kind-create-cluster build-image kind-load-image ## Manage kind cluster

kind-delete-cluster: ## Delete the kind cluster
	kind delete clusters $(KIND_CLUSTER_NAME) || true

kind-create-cluster: ## Create a kind cluster
	kind create cluster --name $(KIND_CLUSTER_NAME) --config manifests/kind-config.yaml --image kindest/node:v$(KIND_NODE_VERSION)


kind-load-image: ## Load docker image into the kind cluster
	kind load docker-image disable-automount-default-sa-controller:1.0.0 --name $(KIND_CLUSTER_NAME)

build-image: ## Build the docker image
	docker build . --tag=disable-automount-default-sa-controller:1.0.0

tests: test ## Run tests

clean-envtest: ## Clean up environment for testing
	rm -rfv ~/.local/share/kubebuilder-envtest

apply-manifest: ## Apply k8s manifests
	kubectl apply -f manifests/deployment.yaml

delete-manifest: ## Delete k8s manifests
	kubectl delete -f manifests/deployment.yaml || true

logs: ## View logs of the controller
	kubectl logs -f -n disable-automount-default-sa-controller -l app=controller

deploy: delete-manifest apply-manifest ## Deploy the application


run-in-kind-cluster: kind delete-manifest apply-manifest logs

