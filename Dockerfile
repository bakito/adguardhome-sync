FROM golang:1.26-alpine AS builder

WORKDIR /go/src/app

RUN apk add --no-cache upx ca-certificates tzdata

ENV CGO_ENABLED=0 \
  GOOS=linux

# Copy only module files first to maximize layer caching for deps.
COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Now copy the rest of the source code.
COPY . .

# Build args should be as late as possible so they don't bust the deps cache.
ARG VERSION=main
ARG BUILD="N/A"

RUN go build -a -installsuffix cgo \
      -ldflags="-w -s -X github.com/bakito/adguardhome-sync/version.Version=${VERSION} -X github.com/bakito/adguardhome-sync/version.Build=${BUILD}" \
      -o adguardhome-sync .

RUN go version && upx -q adguardhome-sync

# application image
FROM scratch
ARG VERSION=main
ARG BUILD="N/A"
WORKDIR /opt/go

LABEL org.opencontainers.image.title="adguardhome-sync" \
      org.opencontainers.image.description="Sync AdGuard Home configuration between instances" \
      org.opencontainers.image.url="https://github.com/bakito/adguardhome-sync" \
      org.opencontainers.image.source="https://github.com/bakito/adguardhome-sync" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD}" \
      org.opencontainers.image.authors="bakito <github@bakito.ch>"
EXPOSE 8080
ENTRYPOINT ["/opt/go/adguardhome-sync"]
CMD ["run", "--config", "/config/adguardhome-sync.yaml"]
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/ /usr/share/zoneinfo/
COPY --from=builder /go/src/app/adguardhome-sync /opt/go/adguardhome-sync
USER 1001
