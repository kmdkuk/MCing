KIND_VERSION = 0.20.0
KUBERNETES_VERSION := v1.28.9# renovate: kindest/node
KUSTOMIZE_VERSION = 5.2.1

PROJECT_DIR :=$(dir $(shell pwd))
BIN_DIR := ${PROJECT_DIR}/bin
KIND := ${BIN_DIR}/kind
KUSTOMIZE := ${BIN_DIR}/kustomize
KUBECTL := ${BIN_DIR}/kubectl
KUBECONFIG := $(shell pwd)/.kubeconfig
KIND_CONFIG = kind-config.yaml
export KUBECTL KUBECONFIG

.PHONY: setup
setup: $(KIND) $(KUBECTL)

.PHONY: start
start: $(KIND) $(KUBECTL)
	$(KIND) create cluster --name=mcing --config=$(KIND_CONFIG) --image=kindest/node:v$(KUBERNETES_VERSION) --wait 1m
	$(MAKE) load
	$(MAKE) cert-manager
	$(KUBECTL) -n kube-system wait --for=condition=available --timeout=180s --all deployments

.PHONY: test
test: $(KUSTOMIZE)
	KINDTEST=1 BIN_DIR=$(BIN_DIR) \
	go test . -v -timeout=15m -ginkgo.v -ginkgo.fail-fast $(GINKGO_OPT)

.PHONY: bootstrap
bootstrap:
	$(MAKE) test GINKGO_OPT="-ginkgo.focus 'bootstrap'"

.PHONY: load
load: $(KIND)
	$(MAKE) -C $(PROJECT_DIR) build-image tag IMAGE_TAG=e2e
	$(KIND) load docker-image mcing-controller:e2e --name mcing
	$(KIND) load docker-image mcing-init:e2e --name mcing
	$(KIND) load docker-image mcing-agent:e2e --name mcing

.PHONY: cert-manager
cert-manager: $(KUBECTL)
	$(KUBECTL) apply -f https://github.com/jetstack/cert-manager/releases/latest/download/cert-manager.yaml
	$(KUBECTL) -n cert-manager wait --for=condition=available --timeout=180s --all deployments

.PHONY: logs
logs:
	rm -rf logs.tar.gz logs
	$(KIND) export logs --name=mcing ./logs
	tar czf logs.tar.gz logs
	rm -rf logs

.PHONY: stop
stop: $(KIND) 
	$(KIND) delete cluster --name=mcing
	-docker image rm mcing-controller:dev
	-docker image rm mcing-init:dev
	-docker image prune -f

$(KIND):
	mkdir -p ${BIN_DIR}
	curl -sfL -o $@ https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-linux-amd64
	chmod a+x $@

$(KUBECTL):
	mkdir -p ${BIN_DIR}
	curl -sfL -o $@ https://dl.k8s.io/release/v$(KUBERNETES_VERSION)/bin/linux/amd64/kubectl
	chmod a+x $@

$(KUSTOMIZE):
	mkdir -p $(BIN_DIR)
	curl -fsL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv$(KUSTOMIZE_VERSION)/kustomize_v$(KUSTOMIZE_VERSION)_linux_amd64.tar.gz | \
	tar -C $(BIN_DIR) -xzf -
