FROM docker.io/library/golang:1.16 as builder

WORKDIR /go/src/app

RUN apt-get update && apt-get install -y upx

ARG VERSION=main
ARG BUILD="N/A"

ENV GOPROXY=https://goproxy.io \
  GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux

ADD . /go/src/app/

RUN go build -a -installsuffix cgo -ldflags="-w -s -X github.com/bakito/adguardhome-sync/version.Version=${VERSION} -X github.com/bakito/adguardhome-sync/version.Build=${BUILD}" -o adguardhome-sync . \
  && upx -q adguardhome-sync

# application image
FROM scratch
WORKDIR /opt/go

LABEL maintainer="bakito <github@bakito.ch>"
EXPOSE 8080
ENTRYPOINT ["/opt/go/adguardhome-sync"]
CMD ["run", "--config", "/config/adguardhome-sync.yaml"]
COPY --from=builder /go/src/app/adguardhome-sync  /opt/go/adguardhome-sync
USER 1001
