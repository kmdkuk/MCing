# Build the manager binary
FROM golang:1.21 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY .git .git
COPY Makefile Makefile
COPY cmd/ cmd/
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/

# Build
RUN make build

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot as controller
WORKDIR /
COPY --from=builder /workspace/bin/mcing-controller .
USER 65532:65532

ENTRYPOINT ["/mcing-controller"]

FROM gcr.io/distroless/static:nonroot as init
WORKDIR /
COPY --from=builder /workspace/bin/mcing-init .
USER 65532:65532

ENTRYPOINT ["/mcing-init"]
