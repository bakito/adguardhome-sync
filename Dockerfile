# syntax=docker/dockerfile:1.7

# Build stage
# NOTE: adguardhome-sync currently requires Go >= 1.25.5
# Use BUILDPLATFORM for the builder to avoid QEMU slow builds for multi-arch.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH
WORKDIR /src

RUN apk add --no-cache ca-certificates git

# Copy go module files first for better caching
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy the rest
COPY . .

# Build upstream adguardhome-sync (root main package)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/adguardhome-sync .

# Build GUI wrapper
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/adguardhome-sync-gui ./cmd/gui

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && adduser -D -H -u 10001 app
WORKDIR /app
COPY --from=builder /out/adguardhome-sync /usr/local/bin/adguardhome-sync
COPY --from=builder /out/adguardhome-sync-gui /usr/local/bin/adguardhome-sync-gui

ENV CONFIG_PATH=/config/adguardhome-sync.yaml
ENV SYNC_BIN=/usr/local/bin/adguardhome-sync
ENV SYNC_ARGS=""
# Sync API (adguardhome-sync built-in UI) is disabled by default to avoid port clash with this GUI (both default to 8080).
# Enable it by setting api.port in YAML (e.g. 8081) and exposing that port, or force via SYNC_API_PORT.
ENV GUI_BIND=0.0.0.0:8080

# Optional basic auth for the GUI:
#   GUI_USERNAME=admin
#   GUI_PASSWORD=changeme
ENV GUI_USERNAME=""
ENV GUI_PASSWORD=""

VOLUME ["/config"]
EXPOSE 8080

USER app
ENTRYPOINT ["/usr/local/bin/adguardhome-sync-gui"]
