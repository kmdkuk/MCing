# ====================================================================================
# Variables
# ====================================================================================

# KUBERNETES_VERSION refers to the version of kindest/node image.
# It does not strictly need to match the kubectl version in aqua.yaml.
KUBERNETES_VERSION := 1.34.3

PROJECT_DIR :=$(dir $(shell pwd))
KUBECONFIG := $(shell pwd)/.kubeconfig
KIND_CONFIG = kind-config.yaml
export KUBECONFIG

# ====================================================================================
# Targets
# ====================================================================================

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: start
start: ## Create kind cluster
	kind create cluster --name=mcing --config=$(KIND_CONFIG) --image=kindest/node:v$(KUBERNETES_VERSION) --wait 1m
	$(MAKE) load
	$(MAKE) cert-manager
	kubectl -n kube-system wait --for=condition=available --timeout=180s --all deployments

.PHONY: stop
stop: ## Delete kind cluster
	kind delete cluster --name=mcing
	-docker image rm mcing-controller:dev
	-docker image rm mcing-init:dev
	-docker image prune -f

.PHONY: test
test: ## Run e2e tests
	KINDTEST=1 \
	go run github.com/onsi/ginkgo/v2/ginkgo -v --timeout=15m --fail-fast -p $(GINKGO_OPT) .

.PHONY: bootstrap
bootstrap: ## Run bootstrap tests
	$(MAKE) test GINKGO_OPT="--focus 'bootstrap'"

.PHONY: load
load: ## Load images into kind cluster
	$(MAKE) -C $(PROJECT_DIR) build-image tag IMAGE_TAG=e2e
	kind load docker-image mcing-controller:e2e --name mcing
	kind load docker-image mcing-init:e2e --name mcing
	kind load docker-image mcing-agent:e2e --name mcing

.PHONY: cert-manager
cert-manager: ## Install cert-manager
	kubectl apply -f https://github.com/jetstack/cert-manager/releases/latest/download/cert-manager.yaml
	kubectl -n cert-manager wait --for=condition=available --timeout=180s --all deployments

.PHONY: logs
logs: ## Export cluster logs
	rm -rf logs.tar.gz logs
	kind export logs --name=mcing ./logs
	tar czf logs.tar.gz logs
	rm -rf logs

.PHONY: clean
clean: ## Clean up logs
	rm -rf logs.tar.gz logs
