KIND_VERSION = 0.11.1
KUBERNETES_VERSION = 1.21.1

MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BIN_DIR := $(shell dirname ${MAKEFILE_DIR})/bin
KIND := ${BIN_DIR}/kind
KUBECTL := ${BIN_DIR}/kubectl

.PHONY: setup
setup: $(KIND) $(KUBECTL)

$(KIND):
	mkdir -p ${BIN_DIR}
	curl -sfL -o $@ https://github.com/kubernetes-sigs/kind/releases/download/v$(KIND_VERSION)/kind-linux-amd64
	chmod a+x $@

$(KUBECTL):
	mkdir -p ${BIN_DIR}
	curl -sfL -o $@ https://dl.k8s.io/release/v$(KUBERNETES_VERSION)/bin/linux/amd64/kubectl
	chmod a+x $@
