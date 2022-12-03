.PHONY: setup-envtest build build-binary tests kind-create-cluster
.DEFAULT_GOAL := all

setup-envtest:
	./setup-envtest.sh

build: test build-binary build-image

build-binary:
	go build -o controller main.go

namespace:
	kubectl delete ns sa-controller --wait=true || true
	kubectl create ns sa-controller

run-binary: build-binary namespace
	CONTROLLER_NAMESPACE=sa-controller ./controller

run: tests build-binary delete-manifest

test:
	SKIP_FETCH_TOOLS=1 ACK_GINKGO_DEPRECATIONS=1.16.5 KUBEBUILDER_ASSETS=~/envtest-binaries/kubebuilder/bin go test -race -v ./... ./controllers/... -count=1 -args -ginkgo.v

kind: kind-delete-cluster kind-create-cluster build-image kind-load-image

kind-delete-cluster:
	kind delete clusters demo || true

kind-create-cluster:
	kind create cluster --name demo --config manifests/kind-config.yaml --image kindest/node:v1.25.0

kind-load-image:
	kind load docker-image disable-automount-default-sa-controller:1.0.0  --name demo

build-image:
	docker build . --tag=disable-automount-default-sa-controller:1.0.0 --no-cache

all: tests build-image build-binary

run: build test build-binary run-binary

tests: clean-envtest setup-envtest test

clean-envtest:
	rm -rfv ~/envtest-binaries/kubebuilder/bin

apply-manifest:
	kubectl apply -f manifests/deployment.yaml

delete-manifest:
	@kubectl delete -f manifests/deployment.yaml || true

logs:
	kubectl logs -f -n disable-automount-default-sa-controller-ns -l app=controller

deploy: delete-manifest apply-manifest

