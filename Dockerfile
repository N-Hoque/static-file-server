################################################################################
## GO BUILDER
################################################################################
FROM golang:1.25.0 as builder

ENV VERSION 1.8.12
ENV CGO_ENABLED 0
ENV BUILD_DIR /build

RUN mkdir -p ${BUILD_DIR}
WORKDIR ${BUILD_DIR}

COPY go.* ./
RUN go mod download
COPY . .

RUN go build -a -tags netgo -installsuffix netgo -ldflags "-s -w -X github.com/N-Hoque/static-file-server/pkg/cli/version.version=${VERSION}" -o /serve /build

################################################################################
## DEPLOYMENT CONTAINER
################################################################################
FROM scratch

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
