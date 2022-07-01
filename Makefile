# Tool versions
CTRL_TOOLS_VERSION=0.9.0
CTRL_RUNTIME_VERSION := $(shell awk '/sigs.k8s.io\/controller-runtime/ {print substr($$2, 2)}' go.mod)
ENVTEST_K8S_VERSION=1.24.2
KUSTOMIZE_VERSION = 4.1.3
CRD_TO_MARKDOWN_VERSION = 0.0.3
MDBOOK_VERSION = 0.4.9

# Test tools
BIN_DIR := $(shell pwd)/bin
STATICCHECK := $(BIN_DIR)/staticcheck
NILERR := $(BIN_DIR)/nilerr

# Set the shell used to bash for better error handling.
SHELL = /usr/bin/env bash
.SHELLFLAGS = -e -o pipefail -c

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS = "crd:crdVersions=v1,maxDescLen=220"

IMAGE_PREFIX :=
IMAGE_TAG := latest

BUILD_FILES = $(shell go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}}\
{{end}}' ./...)

VERSION := $(shell git describe --tags $(shell git rev-list --tags --max-count=1))
DATE_FMT = +%Y-%m-%d
ifdef SOURCE_DATE_EPOCH
    BUILD_DATE ?= $(shell date -u -d "@$(SOURCE_DATE_EPOCH)" "$(DATE_FMT)" 2>/dev/null || date -u -r "$(SOURCE_DATE_EPOCH)" "$(DATE_FMT)" 2>/dev/null || date -u "$(DATE_FMT)")
else
    BUILD_DATE ?= $(shell date "$(DATE_FMT)")
endif

REVISION := $(shell git rev-parse --short HEAD)

GO_LDFLAGS := -X github.com/kmdkuk/mcing/pkg/version.Revision=$(REVISION) $(GO_LDFLAGS)
GO_LDFLAGS := -X github.com/kmdkuk/mcing/pkg/version.BuildDate=$(BUILD_DATE) $(GO_LDFLAGS)
DEV_LDFLAGS := $(GO_LDFLAGS)
GO_LDFLAGS := -X github.com/kmdkuk/mcing/pkg/version.Version=$(VERSION) $(GO_LDFLAGS)

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

clean:
	rm -rf build
	rm -rf bin

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: apidoc
apidoc: crd-to-markdown $(wildcard api/*/*_types.go)
	$(CRD_TO_MARKDOWN) --links docs/links.csv -f api/v1alpha1/minecraft_types.go -n Minecraft > docs/crd_minecraft.md

.PHONY: book
book: mdbook
	rm -rf docs/book
	cd docs; $(MDBOOK) build

.PHONY: check-generate
check-generate:
	$(MAKE) manifests generate apidoc
	git diff --exit-code --name-only

.PHONY: envtest
envtest: setup-envtest
	source <($(SETUP_ENVTEST) use -p env); \
		export DEBUG_CONTROLLER=1; \
		go test -v -count 1 -race ./controllers -ginkgo.progress -ginkgo.v -ginkgo.fail-fast
	source <($(SETUP_ENVTEST) use -p env); \
		go test -v -count 1 -race ./api/... -ginkgo.progress -ginkgo.v

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet setup-envtest test-tools ## Run tests.
	$(STATICCHECK) ./...
	KUBEBUILDER_ASSETS="$(shell $(SETUP_ENVTEST) --arch=amd64 use $(ENVTEST_K8S_VERSION) -p path)" go test -v ./api/... -coverprofile cover.out
	KUBEBUILDER_ASSETS="$(shell $(SETUP_ENVTEST) --arch=amd64 use $(ENVTEST_K8S_VERSION) -p path)" go test -v ./controllers/... -coverprofile cover.out
	KUBEBUILDER_ASSETS="$(shell $(SETUP_ENVTEST) --arch=amd64 use $(ENVTEST_K8S_VERSION) -p path)" go test -v ./pkg/... -coverprofile cover.out

##@ Build

.PHONY: release-build
release-build: kustomize
	mkdir -p build
	$(KUSTOMIZE) build . > build/install.yaml
	$(KUSTOMIZE) build config/samples > build/minecraft-sample.yaml

build: $(BUILD_FILES)
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(GO_LDFLAGS)" -a -o bin/mcing-controller cmd/mcing-controller/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(GO_LDFLAGS)" -a -o bin/mcing-init cmd/mcing-init/main.go

build-image:
	docker build --target controller -t mcing-controller:dev .
	docker build --target init -t mcing-init:dev .

tag:
	docker tag mcing-controller:dev $(IMAGE_PREFIX)mcing-controller:$(IMAGE_TAG)
	docker tag mcing-init:dev $(IMAGE_PREFIX)mcing-init:$(IMAGE_TAG)

push:
	docker push $(IMAGE_PREFIX)mcing-controller:$(IMAGE_TAG)
	docker push $(IMAGE_PREFIX)mcing-init:$(IMAGE_TAG)

##@ Tools

CONTROLLER_GEN := $(BIN_DIR)/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CTRL_TOOLS_VERSION))

SETUP_ENVTEST := $(BIN_DIR)/setup-envtest
.PHONY: setup-envtest
setup-envtest: $(SETUP_ENVTEST)

$(SETUP_ENVTEST): ## Download setup-envtest locally if necessary
	# see https://github.com/kubernetes-sigs/controller-runtime/tree/master/tools/setup-envtest
	GOBIN=$(BIN_DIR) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

KUSTOMIZE := $(BIN_DIR)/kustomize
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.

$(KUSTOMIZE):
	mkdir -p $(BIN_DIR)
	curl -fsL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv$(KUSTOMIZE_VERSION)/kustomize_v$(KUSTOMIZE_VERSION)_linux_amd64.tar.gz | \
	tar -C $(BIN_DIR) -xzf -

CRD_TO_MARKDOWN := $(BIN_DIR)/crd-to-markdown
.PHONY: crd-to-markdown
crd-to-markdown: ## Download crd-to-markdown locally if necessary.
	$(call go-get-tool,$(CRD_TO_MARKDOWN),github.com/clamoriniere/crd-to-markdown@v$(CRD_TO_MARKDOWN_VERSION))

MDBOOK := $(BIN_DIR)/mdbook
.PHONY: mdbook
mdbook: ## Donwload mdbook locally if necessary
	mkdir -p $(BIN_DIR)
	curl -fsL https://github.com/rust-lang/mdBook/releases/download/v$(MDBOOK_VERSION)/mdbook-v$(MDBOOK_VERSION)-x86_64-unknown-linux-gnu.tar.gz | tar -C $(BIN_DIR) -xzf -

# go-get-tool will 'go get' any package $2 and install it to $1.
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
echo "Downloading $(2)" ;\
GOBIN=$(BIN_DIR) go install $(2) ;\
}
endef

.PHONY: test-tools
test-tools: $(STATICCHECK)

$(STATICCHECK):
	mkdir -p $(BIN_DIR)
	GOBIN=$(BIN_DIR) go install honnef.co/go/tools/cmd/staticcheck@latest
