# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot AS controller
LABEL org.opencontainers.image.source="https://github.com/kmdkuk/MCing"
WORKDIR /
COPY LICENSE /
COPY mcing-controller /
USER 65532:65532

ENTRYPOINT ["/mcing-controller"]

FROM gcr.io/distroless/static:nonroot AS init
LABEL org.opencontainers.image.source="https://github.com/kmdkuk/MCing"
WORKDIR /
COPY LICENSE /
COPY mcing-init /
USER 1000:1000

ENTRYPOINT ["/mcing-init"]

FROM gcr.io/distroless/static:nonroot AS agent
LABEL org.opencontainers.image.source="https://github.com/kmdkuk/MCing"
WORKDIR /
COPY LICENSE /
COPY mcing-agent /
USER 1000:1000

ENTRYPOINT ["/mcing-agent"]
