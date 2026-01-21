# ====================================================================================
# Variables
# ====================================================================================

# Tool versions
CTRL_RUNTIME_VERSION := $(shell awk '/sigs.k8s.io\/controller-runtime/ {print substr($$2, 2)}' go.mod)
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.34.1
KUBERNETES_VERSION = 1.34.3

# Image settings
IMAGE_PREFIX :=
IMAGE_TAG := latest
CONTROLLER_IMG ?= mcing-controller:dev
INIT_IMG ?= mcing-init:dev
AGENT_IMG ?= mcing-agent:dev

# Build settings
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

CONTAINER_TOOL ?= docker

BUILD_FILES = $(shell go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}}\
{{end}}' ./...)

# Version info for build
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


# ====================================================================================
# Targets
# ====================================================================================

.PHONY: all
all: build

.PHONY: clean
clean: ## Clean up local bin directory
	rm -rf bin/

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint linter & yamllint
	golangci-lint run

.PHONY: lint-fix
lint-fix: ## Run golangci-lint linter and perform fixes
	golangci-lint run --fix

.PHONY: test
test: manifests generate fmt vet lint ## Run tests.
	KUBEBUILDER_ASSETS="$(shell setup-envtest use $(ENVTEST_K8S_VERSION) -p path)" go test ./api/... -coverprofile cover.out
	KUBEBUILDER_ASSETS="$(shell setup-envtest use $(ENVTEST_K8S_VERSION) -p path)" go test ./internal/... -coverprofile cover.out
	KUBEBUILDER_ASSETS="$(shell setup-envtest use $(ENVTEST_K8S_VERSION) -p path)" go test ./pkg/... -coverprofile cover.out

.PHONY: start
start: ## launc mcing-dev cluster
	ctlptl apply -f ./cluster.yaml
	kubectl apply -f https://github.com/jetstack/cert-manager/releases/latest/download/cert-manager.yaml
	kubectl -n cert-manager wait --for=condition=available --timeout=180s --all deployments

.PHONY: stop
stop: ## stop mcing-dev cluster
	ctlptl delete -f ./cluster.yaml

.PHONY: manifests
manifests: ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: proto
proto: pkg/proto/agentrpc.pb.go pkg/proto/agentrpc_grpc.pb.go docs/agentrpc.md ## Generate proto files

pkg/proto/agentrpc.pb.go: pkg/proto/agentrpc.proto
	protoc -I=. --go_out=module=github.com/kmdkuk/mcing:. $<

pkg/proto/agentrpc_grpc.pb.go: pkg/proto/agentrpc.proto
	protoc -I=. --go-grpc_out=module=github.com/kmdkuk/mcing:. $<

docs/agentrpc.md: pkg/proto/agentrpc.proto
	protoc -I=. --doc_out=docs --doc_opt=markdown,$@ $<

.PHONY: apidoc
apidoc: $(wildcard api/*/*_types.go) ## Generate API docs
	crd-to-markdown --links docs/links.csv -f api/v1alpha1/minecraft_types.go -n Minecraft > docs/crd_minecraft.md

.PHONY: book
book: ## Generate book
	rm -rf docs/book
	cd docs; mdbook build

.PHONY: check-generate
check-generate: ## Check generated files
	$(MAKE) manifests generate apidoc proto
	git diff --exit-code --name-only

##@ Build

.PHONY: build
build: fmt vet $(BUILD_FILES) build-controller build-init build-agent ## Build manager binary.

.PHONY: build-controller
build-controller: fmt vet $(BUILD_FILES) ## Build manager binary.
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(GO_LDFLAGS)" -a -o mcing-controller cmd/mcing-controller/main.go

.PHONY: build-init
build-init: fmt vet $(BUILD_FILES) ## Build manager binary.
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(GO_LDFLAGS)" -a -o mcing-init cmd/mcing-init/main.go

.PHONY: build-agent
build-agent: fmt vet $(BUILD_FILES) ## Build manager binary.
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(GO_LDFLAGS)" -a -o mcing-agent cmd/mcing-agent/main.go

.PHONY: release-build
release-build: ## Build release artifacts
	kustomize build . > install.yaml
	kustomize build config/samples > minecraft-sample.yaml

.PHONY: init-buildx
init-buildx:
	- docker buildx create --name project-v3-builder --driver docker-container --use
	docker buildx inspect --bootstrap

.PHONY: release
release: init-buildx ## Run goreleaser release
	goreleaser release --clean

.PHONY: dry-run-release
dry-run-release: init-buildx ## Run goreleaser release in dry-run mode
	goreleaser release --snapshot --clean --skip=publish

.PHONY: build-image
build-image: build build-image-controller build-image-init build-image-agent ## Build docker image with the manager.

.PHONY: build-image-controller
build-image-controller: build-controller ## Build docker image with the manager.
	$(CONTAINER_TOOL) build --target controller -t ${CONTROLLER_IMG} .

.PHONY: build-image-init
build-image-init: build-init ## Build docker image with the manager.
	$(CONTAINER_TOOL) build --target init -t ${INIT_IMG} .

.PHONY: build-image-agent
build-image-agent: build-agent ## Build docker image with the manager.
	$(CONTAINER_TOOL) build --target agent -t ${AGENT_IMG} .

.PHONY: tag
tag: build tag-controller tag-init tag-agent ## Tag docker image with the manager.

.PHONY: tag-controller
tag-controller: build-controller
	$(CONTAINER_TOOL) tag ${CONTROLLER_IMG} $(IMAGE_PREFIX)mcing-controller:$(IMAGE_TAG)

.PHONY: tag-init
tag-init: build-init
	$(CONTAINER_TOOL) tag ${INIT_IMG} $(IMAGE_PREFIX)mcing-init:$(IMAGE_TAG)

.PHONY: tag-agent
tag-agent: build-agent
	$(CONTAINER_TOOL) tag ${AGENT_IMG} $(IMAGE_PREFIX)mcing-agent:$(IMAGE_TAG)

.PHONY: push
push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push $(IMAGE_PREFIX)mcing-controller:$(IMAGE_TAG)
	$(CONTAINER_TOOL) push $(IMAGE_PREFIX)mcing-init:$(IMAGE_TAG)
	$(CONTAINER_TOOL) push $(IMAGE_PREFIX)mcing-agent:$(IMAGE_TAG)

PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name project-v3-builder
	$(CONTAINER_TOOL) buildx use project-v3-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm project-v3-builder
	rm Dockerfile.cross

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	(kustomize build config/crd | kubectl replace -f -) || (kustomize build config/crd | kubectl create -f -)

.PHONY: uninstall
uninstall: manifests ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	kustomize build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply --server-side -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	kustomize build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -
