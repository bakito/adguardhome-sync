FROM golang:1.23-alpine AS builder

WORKDIR /go/src/app

RUN apk update && apk add upx ca-certificates tzdata

ARG VERSION=main
ARG BUILD="N/A"

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux

COPY . /go/src/app/


RUN go build -a -installsuffix cgo -ldflags="-w -s -X github.com/bakito/adguardhome-sync/version.Version=${VERSION} -X github.com/bakito/adguardhome-sync/version.Build=${BUILD}" -o adguardhome-sync .

RUN go version && upx -q adguardhome-sync

# application image
FROM scratch
WORKDIR /opt/go

LABEL maintainer="bakito <github@bakito.ch>"
EXPOSE 8080
ENTRYPOINT ["/opt/go/adguardhome-sync"]
CMD ["run", "--config", "/config/adguardhome-sync.yaml"]
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/ /usr/share/zoneinfo/
COPY --from=builder /go/src/app/adguardhome-sync  /opt/go/adguardhome-sync
USER 1001
