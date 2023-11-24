load('ext://restart_process', 'docker_build_with_restart')

CONTROLLER_DOCKERFILE = '''FROM golang:alpine
WORKDIR /
COPY ./bin/mcing-controller /
CMD ["/mcing-controller"]
'''

INIT_DOCKERFILE = '''FROM golang:alpine
WORKDIR /
COPY ./bin/mcing-init /
CMD ["/mcing-init"]
'''

AGENT_DOCKERFILE = '''FROM golang:alpine
WORKDIR /
COPY ./bin/mcing-agent /
CMD ["/mcing-agent"]
'''


def manifests():
    return 'make manifests;'

def generate():
    return 'make generate;'

def apidoc():
    return 'make apidoc;'

def controller_binary():
    return 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/mcing-controller cmd/mcing-controller/main.go'

def init_binary():
    return 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/mcing-init cmd/mcing-init/main.go'

def agent_binary():
    return 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/mcing-agent cmd/mcing-agent/main.go'

# Generate manifests and go files
local_resource('make manifests', manifests(), deps=["api", "controllers"], ignore=['*/*/zz_generated.deepcopy.go'])
local_resource('make generate', generate(), deps=["api"], ignore=['*/*/zz_generated.deepcopy.go'])
local_resource('make apidoc', apidoc(), deps=["api"], ignore=['*/*/zz_generated.deepcopy.go'])

# Deploy CRD
local_resource(
    'CRD', manifests() + 'kustomize build config/crd | kubectl apply --server-side --field-manager=tilt -f -', deps=["api"],
    ignore=['*/*/zz_generated.deepcopy.go'])

# Deploy manager
watch_file('./config/')
k8s_yaml(kustomize('./config/dev'))

local_resource(
    'Watch & Compile (mcing controller)', generate() + controller_binary(), deps=['internal/controller', 'api', 'pkg', 'cmd/mcing-controller'],
    ignore=['*/*/zz_generated.deepcopy.go'])

local_resource('Watch & Compile (mcing-init)', init_binary(), deps=['pkg', 'cmd/mcing-init'])

local_resource('Watch & Compile (mcing-agent)', init_binary(), deps=['pkg', 'cmd/mcing-agent'])

docker_build_with_restart(
    'ghcr.io/kmdkuk/mcing-controller:latest', '.',
    dockerfile_contents=CONTROLLER_DOCKERFILE,
    entrypoint=['/mcing-controller'],
    only=['./bin/mcing-controller'],
    live_update=[
        sync('./bin/mcing-controller', '/mcing-controller'),
    ]
)

local_resource('Build & Load (mcing-init)',
    'make build-image tag IMAGE_PREFIX=ghcr.io/kmdkuk/ IMAGE_TAG=e2e; kind load docker-image ghcr.io/kmdkuk/mcing-init:e2e --name mcing-dev',
    deps=["Dockerfile", "./bin/mcing-init"])

local_resource('Build & Load (mcing-agent)',
    'make build-image tag IMAGE_PREFIX=ghcr.io/kmdkuk/ IMAGE_TAG=e2e; kind load docker-image ghcr.io/kmdkuk/mcing-agent:e2e --name mcing-dev',
    deps=["Dockerfile", "./bin/mcing-agent"])
