# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot AS controller
LABEL org.opencontainers.image.source="https://github.com/kmdkuk/MCing"
WORKDIR /
COPY LICENSE /
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/mcing-controller /
USER 65532:65532

ENTRYPOINT ["/mcing-controller"]

FROM ubuntu:24.04 AS lazymc
WORKDIR /
RUN apt update && apt install -y curl ca-certificates
ARG LAZYMC_VERSION="v0.2.11"
ARG BASE_URL="https://github.com/timvisee/lazymc/releases/download"
ARG FILE_NAME="lazymc-${LAZYMC_VERSION}-linux-x64-static"
ARG TARGET_URL="${BASE_URL}/${LAZYMC_VERSION}/${FILE_NAME}"
RUN --mount=type=secret,id=github_token <<EOF
if [ -f /run/secrets/github_token ]; then
    echo "Token found.";
    AUTH_HEADER="Authorization: Bearer $(cat /run/secrets/github_token)";
else
    echo "No token found.";
    AUTH_HEADER="User-Agent: Mozilla/5.0";
fi;
if [ -n "$AUTH_HEADER" ] && [ "$AUTH_HEADER" != "User-Agent: Mozilla/5.0" ]; then
    curl -L -H "$AUTH_HEADER" -o /lazymc $TARGET_URL;
else
    curl -L -o /lazymc $TARGET_URL;
fi;
chmod +x /lazymc
EOF

FROM gcr.io/distroless/static:nonroot AS init
LABEL org.opencontainers.image.source="https://github.com/kmdkuk/MCing"
WORKDIR /
COPY LICENSE /
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/mcing-init /
USER 1000:1000

COPY --from=lazymc /lazymc /

ENTRYPOINT ["/mcing-init"]

FROM gcr.io/distroless/static:nonroot AS agent
LABEL org.opencontainers.image.source="https://github.com/kmdkuk/MCing"
WORKDIR /
COPY LICENSE /
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/mcing-agent /
USER 1000:1000

ENTRYPOINT ["/mcing-agent"]
