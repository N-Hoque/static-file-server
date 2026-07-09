# syntax=docker/dockerfile:1

################################################################################
## GO BUILDER
################################################################################
FROM --platform=$BUILDPLATFORM golang:1.26.5 AS builder

ARG VERSION
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG TARGETVARIANT

WORKDIR /build
COPY go.* ./
RUN go mod download
COPY . .

# Build for the target platform using Go's native cross-compilation.
# BUILDPLATFORM keeps the builder on the native host so no QEMU emulation
# is needed in this stage.  TARGETVARIANT carries "v7" for linux/arm/v7;
# strip the leading "v" to produce the GOARM value that Go expects.
RUN --mount=type=cache,target=/root/.cache/go-build \
    set -e; \
    GOARM="${TARGETVARIANT#v}"; \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${GOARM} \
    go build \
        -ldflags "-s -w -X github.com/N-Hoque/static-file-server/pkg/cli/version.version=${VERSION}" \
        -o /serve cmd/serve/main.go

################################################################################
## DEPLOYMENT CONTAINER
################################################################################
FROM scratch

ARG VERSION
EXPOSE 8080

COPY --from=builder /serve /serve

ENTRYPOINT ["/serve"]
CMD []

# Metadata
LABEL life.apets.vendor="N-Hoque" \
      life.apets.url="https://github.com/N-Hoque/static-file-server" \
      life.apets.name="Static File Server" \
      life.apets.description="A tiny static file server (forked from Halverneus)" \
      life.apets.version="v${VERSION}" \
      life.apets.schema-version="1.0"
