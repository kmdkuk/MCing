load('ext://restart_process', 'docker_build_with_restart')

CONTROLLER_DOCKERFILE = '''FROM golang:alpine
WORKDIR /
COPY ./bin/mcing-controller /
CMD ["/mcing-controller"]
'''


def manifests():
    return 'make manifests;'

def generate():
    return 'make generate;'

def apidoc():
    return 'make apidoc;'

def controller_binary():
    return 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/mcing-controller cmd/mcing-controller/main.go'

# Generate manifests and go files
local_resource('make manifests', manifests(), deps=["api", "controllers"], ignore=['*/*/zz_generated.deepcopy.go'])
local_resource('make generate', generate(), deps=["api"], ignore=['*/*/zz_generated.deepcopy.go'])
local_resource('make apidoc', apidoc(), deps=["api"], ignore=['*/*/zz_generated.deepcopy.go'])

# Deploy CRD
local_resource(
    'CRD', 'make install', deps=["api"],
    ignore=['*/*/zz_generated.deepcopy.go'])

# Deploy manager
watch_file('./config/')
k8s_yaml(kustomize('./config/dev'))

# Deploy sample minecraft resource
local_resource('Sample YAML', 'kustomize build ./config/samples | kubectl apply -f -', deps=["./config/samples"], resource_deps=["mcing-controller-manager"])

local_resource(
    'Watch & Compile (mcing controller)', generate() + controller_binary(), deps=['internal', 'api', 'pkg', 'cmd/mcing-controller'],
    ignore=['*/*/zz_generated.deepcopy.go'])

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
    'make build-image-init tag-init IMAGE_PREFIX=ghcr.io/kmdkuk/ IMAGE_TAG=e2e; kind load docker-image ghcr.io/kmdkuk/mcing-init:e2e --name mcing-dev',
    deps=["Dockerfile", 'pkg', 'cmd/mcing-init'])

local_resource('Build & Load (mcing-agent)',
    'make build-image-agent tag-agent IMAGE_PREFIX=ghcr.io/kmdkuk/ IMAGE_TAG=e2e; kind load docker-image ghcr.io/kmdkuk/mcing-agent:e2e --name mcing-dev',
    deps=["Dockerfile", 'pkg', 'cmd/mcing-agent'])
