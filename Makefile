# Include toolbox tasks
include ./.toolbox.mk

# Run go lint against code
lint: tb.golangci-lint
	$(TB_GOLANGCI_LINT) run --fix

# Run go mod tidy
tidy:
	go mod tidy

generate: tb.controller-gen
	@mkdir -p ./tmp
	@touch ./tmp/deepcopy-gen-boilerplate.go.txt
	$(TB_CONTROLLER_GEN) paths=./pkg/types object

fmt: tb.golines tb.gofumpt
	$(TB_GOLINES) --base-formatter="$(TB_GOFUMPT)" --max-len=120 --write-output .

# Run tests
test: generate fmt lint test-ci

fuzz:
	 go test -fuzz=FuzzMask -v ./pkg/types/ -fuzztime=60s

# Run ci tests
test-ci: mocks tidy tb.ginkgo
	$(TB_GINKGO) --cover --coverprofile coverage.out.tmp ./...
	cat coverage.out.tmp | grep -v "_generated.go" > coverage.out
	go tool cover -func=coverage.out

mocks: tb.mockgen
	$(TB_MOCKGEN) -package client -destination pkg/mocks/client/mock.go github.com/bakito/adguardhome-sync/pkg/client Client
	$(TB_MOCKGEN) -package client -destination pkg/mocks/flags/mock.go github.com/bakito/adguardhome-sync/pkg/config Flags

release: tb.semver tb.goreleaser
	@version=$$($(TB_SEMVER)); \
	git tag -s $$version -m"Release $$version"
	$(TB_GORELEASER) --clean

test-release: tb.goreleaser
	$(TB_GORELEASER) --skip=publish --snapshot --clean

start-replica:
	docker rm -f adguardhome-replica
	docker run --pull always --name adguardhome-replica -p 9091:3000 --rm adguard/adguardhome:latest
#	docker run --pull always --name adguardhome-replica -p 9090:80 -p 9091:3000 --rm adguard/adguardhome:v0.107.13

copy-replica-config:
	docker cp adguardhome-replica:/opt/adguardhome/conf/AdGuardHome.yaml tmp/AdGuardHome.yaml

start-replica2:
	docker rm -f adguardhome-replica2
	docker run --pull always --name adguardhome-replica2 -p 9093:3000 --rm adguard/adguardhome:latest
#	docker run --pull always --name adguardhome-replica -p 9090:80 -p 9091:3000 --rm adguard/adguardhome:v0.107.13

check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))

build-image:
	$(call check_defined, AGH_SYNC_VERSION)
	docker build --build-arg VERSION=${AGH_SYNC_VERSION} --build-arg BUILD=$(shell date -u +'%Y-%m-%dT%H:%M:%S.%3NZ') --name adgardhome-replica -t ghcr.io/bakito/adguardhome-sync:${AGH_SYNC_VERSION} .

kind-create:
	kind delete cluster
	kind create  cluster

kind-test:
	@./testdata/e2e/bin/install-chart.sh

# renovate: packageName=AdguardTeam/AdGuardHome
ADGUARD_HOME_VERSION ?= v0.107.61

model: tb.oapi-codegen
	@mkdir -p tmp
	go run openapi/main.go $(ADGUARD_HOME_VERSION)
	$(TB_OAPI_CODEGEN) -package model -generate types,client -config .oapi-codegen.yaml tmp/schema.yaml > pkg/client/model/model_generated.go

model-diff:
	go run openapi/main.go $(ADGUARD_HOME_VERSION)
	go run openapi/main.go
	diff tmp/schema.yaml tmp/schema-master.yaml

zellij:
	zellij -l ./testdata/test-layout.kdl
