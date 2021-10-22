KIND_VERSION = 0.11.1
KUBERNETES_VERSION = 1.21.1
CERT_MANAGER_VERSION = 1.4.0

BIN_DIR := $(dir $(shell pwd))/bin
KIND := ${BIN_DIR}/kind
KUBECTL := ${BIN_DIR}/kubectl
KUBECONFIG := $(shell pwd)/.kubeconfig
KIND_CONFIG = kind-config.yaml
export KUBECTL KUBECONFIG

.PHONY: setup
setup: $(KIND) $(KUBECTL)

.PHONY: start
start: $(KIND) $(KUBECTL)
	$(KIND) create cluster --name=mcing --config=$(KIND_CONFIG) --image=kindest/node:v$(KUBERNETES_VERSION) --wait 1m
	cd ..; docker build --no-cache -t mcing:dev .
	$(KIND) load docker-image mcing:dev --name=mcing
	$(KUBECTL) apply -f https://github.com/jetstack/cert-manager/releases/download/v$(CERT_MANAGER_VERSION)/cert-manager.yaml
	$(KUBECTL) -n cert-manager wait --for=condition=available --timeout=180s --all deployments
	$(KUBECTL) -n kube-system wait --for=condition=available --timeout=180s --all deployments

.PHONY: stop
stop: $(KIND) 
	$(KIND) delete cluster --name=mcing
	-docker image rm mcing:dev
	-docker image prune -f

$(KIND):
	mkdir -p ${BIN_DIR}
	curl -sfL -o $@ https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-linux-amd64
	chmod a+x $@

$(KUBECTL):
	mkdir -p ${BIN_DIR}
	curl -sfL -o $@ https://dl.k8s.io/release/v$(KUBERNETES_VERSION)/bin/linux/amd64/kubectl
	chmod a+x $@
