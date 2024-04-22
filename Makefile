# Run go lint against code
lint: golangci-lint
	$(GOLANGCI_LINT) run --fix

# Run go mod tidy
tidy:
	go mod tidy

generate: deepcopy-gen
	@mkdir -p ./tmp
	@touch ./tmp/deepcopy-gen-boilerplate.go.txt
	$(DEEPCOPY_GEN) -h ./tmp/deepcopy-gen-boilerplate.go.txt -i ./pkg/types

# Run tests
test: generate lint test-ci

# Run ci tests
test-ci: mocks tidy ginkgo
	$(GINKGO) --cover --coverprofile coverage.out.tmp ./...
	cat coverage.out.tmp | grep -v "_generated.go" > coverage.out
	go tool cover -func=coverage.out

mocks: mockgen
	$(MOCKGEN) -package client -destination pkg/mocks/client/mock.go github.com/bakito/adguardhome-sync/pkg/client Client
	$(MOCKGEN) -package client -destination pkg/mocks/flags/mock.go github.com/bakito/adguardhome-sync/pkg/config Flags

release: semver goreleaser
	@version=$$($(LOCALBIN)/semver); \
	git tag -s $$version -m"Release $$version"
	$(GORELEASER) --clean

test-release: goreleaser
	$(GORELEASER) --skip=publish --snapshot --clean

## toolbox - start
## Current working directory
LOCALDIR ?= $(shell which cygpath > /dev/null 2>&1 && cygpath -m $$(pwd) || pwd)
## Location to install dependencies to
LOCALBIN ?= $(LOCALDIR)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
DEEPCOPY_GEN ?= $(LOCALBIN)/deepcopy-gen
GINKGO ?= $(LOCALBIN)/ginkgo
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
GORELEASER ?= $(LOCALBIN)/goreleaser
MOCKGEN ?= $(LOCALBIN)/mockgen
OAPI_CODEGEN ?= $(LOCALBIN)/oapi-codegen
SEMVER ?= $(LOCALBIN)/semver

## Tool Versions
DEEPCOPY_GEN_VERSION ?= v0.30.0
GINKGO_VERSION ?= v2.17.1
GOLANGCI_LINT_VERSION ?= v1.57.2
GORELEASER_VERSION ?= v1.25.1
MOCKGEN_VERSION ?= v0.4.0
OAPI_CODEGEN_VERSION ?= v2.1.0
SEMVER_VERSION ?= v1.1.3

## Tool Installer
.PHONY: deepcopy-gen
deepcopy-gen: $(DEEPCOPY_GEN) ## Download deepcopy-gen locally if necessary.
$(DEEPCOPY_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/deepcopy-gen || GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/deepcopy-gen@$(DEEPCOPY_GEN_VERSION)
.PHONY: ginkgo
ginkgo: $(GINKGO) ## Download ginkgo locally if necessary.
$(GINKGO): $(LOCALBIN)
	test -s $(LOCALBIN)/ginkgo || GOBIN=$(LOCALBIN) go install github.com/onsi/ginkgo/v2/ginkgo@$(GINKGO_VERSION)
.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	test -s $(LOCALBIN)/golangci-lint || GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
.PHONY: goreleaser
goreleaser: $(GORELEASER) ## Download goreleaser locally if necessary.
$(GORELEASER): $(LOCALBIN)
	test -s $(LOCALBIN)/goreleaser || GOBIN=$(LOCALBIN) go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION)
.PHONY: mockgen
mockgen: $(MOCKGEN) ## Download mockgen locally if necessary.
$(MOCKGEN): $(LOCALBIN)
	test -s $(LOCALBIN)/mockgen || GOBIN=$(LOCALBIN) go install go.uber.org/mock/mockgen@$(MOCKGEN_VERSION)
.PHONY: oapi-codegen
oapi-codegen: $(OAPI_CODEGEN) ## Download oapi-codegen locally if necessary.
$(OAPI_CODEGEN): $(LOCALBIN)
	test -s $(LOCALBIN)/oapi-codegen || GOBIN=$(LOCALBIN) go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION)
.PHONY: semver
semver: $(SEMVER) ## Download semver locally if necessary.
$(SEMVER): $(LOCALBIN)
	test -s $(LOCALBIN)/semver || GOBIN=$(LOCALBIN) go install github.com/bakito/semver@$(SEMVER_VERSION)

## Update Tools
.PHONY: update-toolbox-tools
update-toolbox-tools:
	@rm -f \
		$(LOCALBIN)/deepcopy-gen \
		$(LOCALBIN)/ginkgo \
		$(LOCALBIN)/golangci-lint \
		$(LOCALBIN)/goreleaser \
		$(LOCALBIN)/mockgen \
		$(LOCALBIN)/oapi-codegen \
		$(LOCALBIN)/semver
	toolbox makefile -f $(LOCALDIR)/Makefile \
		k8s.io/code-generator/cmd/deepcopy-gen@github.com/kubernetes/code-generator \
		github.com/onsi/ginkgo/v2/ginkgo \
		github.com/golangci/golangci-lint/cmd/golangci-lint \
		github.com/goreleaser/goreleaser \
		go.uber.org/mock/mockgen@github.com/uber-go/mock \
		github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen \
		github.com/bakito/semver
## toolbox - end

start-replica:
	docker run --pull always --name adguardhome-replica -p 9091:3000 --rm adguard/adguardhome:latest
#	docker run --pull always --name adguardhome-replica -p 9090:80 -p 9091:3000 --rm adguard/adguardhome:v0.107.13

copy-replica-config:
	docker cp adguardhome-replica:/opt/adguardhome/conf/AdGuardHome.yaml tmp/AdGuardHome.yaml

start-replica2:
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

model: oapi-codegen
	@mkdir -p tmp
	go run openapi/main.go v0.107.46
	$(OAPI_CODEGEN) -package model -generate types,client -config .oapi-codegen.yaml tmp/schema.yaml > pkg/client/model/model_generated.go

model-diff:
	go run openapi/main.go v0.107.46
	go run openapi/main.go
	diff tmp/schema.yaml tmp/schema-master.yaml
