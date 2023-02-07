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
test-ci: mocks tidy
	go test ./...  -coverprofile=coverage.out
	go tool cover -func=coverage.out

mocks: mockgen
	$(MOCKGEN) -package client -destination pkg/mocks/client/mock.go github.com/bakito/adguardhome-sync/pkg/client Client

release: semver goreleaser
	@version=$$($(LOCALBIN)/semver);
	git tag -s $$version -m"Release $$version"
	$(GORELEASER) --rm-dist

test-release: goreleaser
	$(GORELEASER) --skip-publish --snapshot --rm-dist

## toolbox - start
## Current working directory
LOCALDIR ?= $(shell which cygpath > /dev/null 2>&1 && cygpath -m $$(pwd) || pwd)
## Location to install dependencies to
LOCALBIN ?= $(LOCALDIR)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
SEMVER ?= $(LOCALBIN)/semver
MOCKGEN ?= $(LOCALBIN)/mockgen
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
GORELEASER ?= $(LOCALBIN)/goreleaser
DEEPCOPY_GEN ?= $(LOCALBIN)/deepcopy-gen

## Tool Versions
SEMVER_VERSION ?= v1.1.3
MOCKGEN_VERSION ?= v1.6.0
GOLANGCI_LINT_VERSION ?= v1.51.1
GORELEASER_VERSION ?= v1.15.1
DEEPCOPY_GEN_VERSION ?= v0.26.1

## Tool Installer
.PHONY: semver
semver: $(SEMVER) ## Download semver locally if necessary.
$(SEMVER): $(LOCALBIN)
	test -s $(LOCALBIN)/semver || GOBIN=$(LOCALBIN) go install github.com/bakito/semver@$(SEMVER_VERSION)
.PHONY: mockgen
mockgen: $(MOCKGEN) ## Download mockgen locally if necessary.
$(MOCKGEN): $(LOCALBIN)
	test -s $(LOCALBIN)/mockgen || GOBIN=$(LOCALBIN) go install github.com/golang/mock/mockgen@$(MOCKGEN_VERSION)
.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	test -s $(LOCALBIN)/golangci-lint || GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
.PHONY: goreleaser
goreleaser: $(GORELEASER) ## Download goreleaser locally if necessary.
$(GORELEASER): $(LOCALBIN)
	test -s $(LOCALBIN)/goreleaser || GOBIN=$(LOCALBIN) go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION)
.PHONY: deepcopy-gen
deepcopy-gen: $(DEEPCOPY_GEN) ## Download deepcopy-gen locally if necessary.
$(DEEPCOPY_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/deepcopy-gen || GOBIN=$(LOCALBIN) go install k8s.io/code-generator/cmd/deepcopy-gen@$(DEEPCOPY_GEN_VERSION)

## Update Tools
.PHONY: update-toolbox-tools
update-toolbox-tools:
	@rm -f \
		$(LOCALBIN)/semver \
		$(LOCALBIN)/mockgen \
		$(LOCALBIN)/golangci-lint \
		$(LOCALBIN)/goreleaser \
		$(LOCALBIN)/deepcopy-gen
	toolbox makefile -f $(LOCALDIR)/Makefile \
		github.com/bakito/semver \
		github.com/golang/mock/mockgen \
		github.com/golangci/golangci-lint/cmd/golangci-lint \
		github.com/goreleaser/goreleaser \
		k8s.io/code-generator/cmd/deepcopy-gen@github.com/kubernetes/code-generator
## toolbox - end

start-replica:
	docker run --pull always --name adguardhome-replica -p 9091:3000 --rm adguard/adguardhome:v0.107.23
#	docker run --pull always --name adguardhome-replica -p 9090:80 -p 9091:3000 --rm adguard/adguardhome:v0.107.13

start-replica2:
	docker run --pull always --name adguardhome-replica2 -p 9093:3000 --rm adguard/adguardhome:v0.107.23
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
