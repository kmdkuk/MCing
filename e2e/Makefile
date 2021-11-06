KIND_VERSION = 0.11.1
KUBERNETES_VERSION = 1.21.2
CERT_MANAGER_VERSION = 1.4.0
KUSTOMIZE_VERSION = 4.1.3

AGENT_VERSION = 0.0.3

PROJECT_DIR :=$(dir $(shell pwd))
BIN_DIR := ${PROJECT_DIR}/bin
KIND := ${BIN_DIR}/kind
KUSTOMIZE := ${BIN_DIR}/kustomize
KUBECTL := ${BIN_DIR}/kubectl
KUBECTL_MCING := ${BIN_DIR}/kubectl-mcing
KUBECONFIG := $(shell pwd)/.kubeconfig
KIND_CONFIG = kind-config.yaml
export KUBECTL KUBECONFIG

.PHONY: setup
setup: $(KIND) $(KUBECTL)

.PHONY: start
start: $(KIND) $(KUBECTL) $(KUBECTL_MCING)
	$(KIND) create cluster --name=mcing --config=$(KIND_CONFIG) --image=kindest/node:v$(KUBERNETES_VERSION) --wait 1m
	$(MAKE) load
	$(MAKE) cert-manager
	$(KUBECTL) -n kube-system wait --for=condition=available --timeout=180s --all deployments

.PHONY: test
test: $(KUSTOMIZE)
	KINDTEST=1 BIN_DIR=$(BIN_DIR) \
	go test . -v -timeout=15m -ginkgo.v -ginkgo.failFast $(GINKGO_OPT)

.PHONY: bootstrap
bootstrap:
	$(MAKE) test GINKGO_OPT="-ginkgo.focus 'bootstrap'"

.PHONY: load
load: $(KIND)
	$(MAKE) -C $(PROJECT_DIR) build-image tag IMAGE_TAG=e2e
	docker pull ghcr.io/kmdkuk/mcing-agent:${AGENT_VERSION}
	$(KIND) load docker-image mcing-controller:e2e --name mcing
	$(KIND) load docker-image mcing-init:e2e --name mcing
	$(KIND) load docker-image ghcr.io/kmdkuk/mcing-agent:${AGENT_VERSION} --name mcing

.PHONY: cert-manager
cert-manager: $(KUBECTL)
	$(KUBECTL) apply -f https://github.com/jetstack/cert-manager/releases/download/v$(CERT_MANAGER_VERSION)/cert-manager.yaml
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

$(KUBECTL_MCING):
	mkdir -p ../bin
	cd ..; GOBIN=$$(pwd)/bin go install ./cmd/kubectl-mcing
