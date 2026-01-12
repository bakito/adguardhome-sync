## toolbox - start
## Generated with https://github.com/bakito/toolbox

## Current working directory
TB_LOCALDIR ?= $(shell which cygpath > /dev/null 2>&1 && cygpath -m $$(pwd) || pwd)
## Location to install dependencies to
TB_LOCALBIN ?= $(TB_LOCALDIR)/bin
$(TB_LOCALBIN):
	if [ ! -e $(TB_LOCALBIN) ]; then mkdir -p $(TB_LOCALBIN); fi

# Helper functions
STRIP_V = $(patsubst v%,%,$(1))

## Tool Binaries
TB_CONTROLLER_GEN ?= $(TB_LOCALBIN)/controller-gen
TB_GINKGO ?= $(TB_LOCALBIN)/ginkgo
TB_GOLANGCI_LINT ?= $(TB_LOCALBIN)/golangci-lint
TB_GORELEASER ?= $(TB_LOCALBIN)/goreleaser
TB_MOCKGEN ?= $(TB_LOCALBIN)/mockgen
TB_OAPI_CODEGEN ?= $(TB_LOCALBIN)/oapi-codegen
TB_SEMVER ?= $(TB_LOCALBIN)/semver
TB_SYFT ?= $(TB_LOCALBIN)/syft

## Tool Versions
# renovate: packageName=github.com/kubernetes-sigs/controller-tools
TB_CONTROLLER_GEN_VERSION ?= v0.20.0
# renovate: packageName=github.com/golangci/golangci-lint/v2
TB_GOLANGCI_LINT_VERSION ?= v2.7.2
TB_GOLANGCI_LINT_VERSION_NUM ?= $(call STRIP_V,$(TB_GOLANGCI_LINT_VERSION))
# renovate: packageName=github.com/goreleaser/goreleaser/v2
TB_GORELEASER_VERSION ?= v2.13.3
TB_GORELEASER_VERSION_NUM ?= $(call STRIP_V,$(TB_GORELEASER_VERSION))
# renovate: packageName=github.com/uber-go/mock
TB_MOCKGEN_VERSION ?= v0.6.0
# renovate: packageName=github.com/oapi-codegen/oapi-codegen/v2
TB_OAPI_CODEGEN_VERSION ?= v2.5.1
# renovate: packageName=github.com/bakito/semver
TB_SEMVER_VERSION ?= v1.1.7
TB_SEMVER_VERSION_NUM ?= $(call STRIP_V,$(TB_SEMVER_VERSION))
# renovate: packageName=github.com/anchore/syft/cmd/syft
TB_SYFT_VERSION ?= v1.39.0
TB_SYFT_VERSION_NUM ?= $(call STRIP_V,$(TB_SYFT_VERSION))

## Tool Installer
.PHONY: tb.controller-gen
tb.controller-gen: ## Download controller-gen locally if necessary.
	@test -s $(TB_CONTROLLER_GEN) || \
		GOBIN=$(TB_LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(TB_CONTROLLER_GEN_VERSION)
.PHONY: tb.ginkgo
tb.ginkgo: ## Download ginkgo locally if necessary.
	@test -s $(TB_GINKGO) || \
		GOBIN=$(TB_LOCALBIN) go install github.com/onsi/ginkgo/v2/ginkgo
.PHONY: tb.golangci-lint
tb.golangci-lint: ## Download golangci-lint locally if necessary.
	@test -s $(TB_GOLANGCI_LINT) && $(TB_GOLANGCI_LINT) --version | grep -q $(TB_GOLANGCI_LINT_VERSION_NUM) || \
		GOBIN=$(TB_LOCALBIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(TB_GOLANGCI_LINT_VERSION)
.PHONY: tb.goreleaser
tb.goreleaser: ## Download goreleaser locally if necessary.
	@test -s $(TB_GORELEASER) && $(TB_GORELEASER) --version | grep -q $(TB_GORELEASER_VERSION_NUM) || \
		GOBIN=$(TB_LOCALBIN) go install github.com/goreleaser/goreleaser/v2@$(TB_GORELEASER_VERSION)
.PHONY: tb.mockgen
tb.mockgen: ## Download mockgen locally if necessary.
	@test -s $(TB_MOCKGEN) || \
		GOBIN=$(TB_LOCALBIN) go install go.uber.org/mock/mockgen@$(TB_MOCKGEN_VERSION)
.PHONY: tb.oapi-codegen
tb.oapi-codegen: ## Download oapi-codegen locally if necessary.
	@test -s $(TB_OAPI_CODEGEN) || \
		GOBIN=$(TB_LOCALBIN) go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(TB_OAPI_CODEGEN_VERSION)
.PHONY: tb.semver
tb.semver: ## Download semver locally if necessary.
	@test -s $(TB_SEMVER) && $(TB_SEMVER) -version | grep -q $(TB_SEMVER_VERSION_NUM) || \
		GOBIN=$(TB_LOCALBIN) go install github.com/bakito/semver@$(TB_SEMVER_VERSION)
.PHONY: tb.syft
tb.syft: ## Download syft locally if necessary.
	@test -s $(TB_SYFT) && $(TB_SYFT) --version | grep -q $(TB_SYFT_VERSION_NUM) || \
		GOBIN=$(TB_LOCALBIN) go install github.com/anchore/syft/cmd/syft@$(TB_SYFT_VERSION)

## Reset Tools
.PHONY: tb.reset
tb.reset:
	@rm -f \
		$(TB_CONTROLLER_GEN) \
		$(TB_GINKGO) \
		$(TB_GOLANGCI_LINT) \
		$(TB_GORELEASER) \
		$(TB_MOCKGEN) \
		$(TB_OAPI_CODEGEN) \
		$(TB_SEMVER) \
		$(TB_SYFT)

## Update Tools
.PHONY: tb.update
tb.update: tb.reset
	toolbox makefile --renovate -f $(TB_LOCALDIR)/Makefile \
		sigs.k8s.io/controller-tools/cmd/controller-gen@github.com/kubernetes-sigs/controller-tools \
		github.com/golangci/golangci-lint/v2/cmd/golangci-lint?--version \
		github.com/goreleaser/goreleaser/v2?--version \
		go.uber.org/mock/mockgen@github.com/uber-go/mock \
		github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		github.com/bakito/semver?-version \
		github.com/anchore/syft/cmd/syft?--version
## toolbox - end
