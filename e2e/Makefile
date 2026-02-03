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
# E2E parallel settings
# ====================================================================================
# E2E_PROCS: explicitly set parallel procs (overrides auto).
# E2E_MEM_PER_PROC_MB: estimated memory per instance, used to cap parallelism.
# E2E_MAX_PROCS: hard cap for auto parallelism.
E2E_PROCS ?=
E2E_MEM_PER_PROC_MB ?= 2048
E2E_MEM_BUFFER_PCT ?= 20
E2E_MAX_PROCS ?=

E2E_CPU_PROCS := $(shell (getconf _NPROCESSORS_ONLN 2>/dev/null || nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 1) | head -n1)
E2E_MEM_PROCS := $(shell \
	mem_bytes=$$(awk '/MemAvailable/ {print $$2*1024}' /proc/meminfo 2>/dev/null); \
	if [ -z "$$mem_bytes" ]; then mem_bytes=$$(awk '/MemTotal/ {print $$2*1024}' /proc/meminfo 2>/dev/null); fi; \
	if [ -z "$$mem_bytes" ]; then mem_bytes=$$(sysctl -n hw.memsize 2>/dev/null); fi; \
	if [ -z "$$mem_bytes" ]; then echo 0; else \
		usable=$$((mem_bytes * (100 - $(E2E_MEM_BUFFER_PCT)) / 100)); \
		echo $$((usable / 1024 / 1024 / $(E2E_MEM_PER_PROC_MB))); \
	fi)
E2E_AUTO_PROCS := $(shell \
	cpu="$(E2E_CPU_PROCS)"; \
	mem="$(E2E_MEM_PROCS)"; \
	if [ "$$mem" -le 0 ]; then p="$$cpu"; \
	elif [ "$$mem" -lt "$$cpu" ]; then p="$$mem"; else p="$$cpu"; fi; \
	if [ "$$p" -lt 1 ]; then p=1; fi; \
	if [ -n "$(E2E_MAX_PROCS)" ] && [ "$(E2E_MAX_PROCS)" -lt "$$p" ]; then p="$(E2E_MAX_PROCS)"; fi; \
	echo $$p)

GINKGO_PARALLEL_FLAG :=
ifneq ($(strip $(E2E_PROCS)),)
GINKGO_PARALLEL_FLAG := --procs=$(E2E_PROCS)
else
GINKGO_PARALLEL_FLAG := --procs=$(E2E_AUTO_PROCS)
endif

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
	$(MAKE) -C $(PROJECT_DIR) build-kubectl-mcing
	KINDTEST=1 BIN_DIR=$(PROJECT_DIR)/bin \
	go run github.com/onsi/ginkgo/v2/ginkgo -v --timeout=15m --fail-fast -p $(GINKGO_PARALLEL_FLAG) $(GINKGO_OPT) .

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
