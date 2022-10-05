# Run go lint against code
lint:
	golangci-lint run --fix

# Run go mod tidy
tidy:
	go mod tidy

generate: deepcopy-gen
	touch ./tmp/deepcopy-gen-boilerplate.go.txt
	deepcopy-gen -h ./tmp/deepcopy-gen-boilerplate.go.txt -i ./pkg/types

# Run tests
test: generate lint test-ci

# Run ci tests
test-ci: mocks tidy
	go test ./...  -coverprofile=coverage.out
	go tool cover -func=coverage.out

mocks: mockgen
	mockgen -package client -destination pkg/mocks/client/mock.go github.com/bakito/adguardhome-sync/pkg/client Client

release: semver
	@version=$$(semver); \
	git tag -s $$version -m"Release $$version"
	goreleaser --rm-dist

test-release:
	goreleaser --skip-publish --snapshot --rm-dist

semver:
ifeq (, $(shell which semver))
 $(shell go install github.com/bakito/semver@latest)
endif

mockgen:
ifeq (, $(shell which mockgen))
 $(shell go install github.com/golang/mock/mockgen@v1.6.0)
endif

deepcopy-gen:
ifeq (, $(shell which deepcopy-gen))
 $(shell go install k8s.io/code-generator/cmd/deepcopy-gen@latest)
endif

start-replica:
	podman run --pull always --name adguardhome-replica -p 9090:80 -p 9091:3000 --rm adguard/adguardhome
#	podman run --pull always --name adguardhome-replica -p 9090:80 -p 9091:3000 --rm adguard/adguardhome:v0.107.13

check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))

build-image:
	$(call check_defined, AGH_SYNC_VERSION)
	podman build --build-arg VERSION=${AGH_SYNC_VERSION} --build-arg BUILD=$(shell date -u +'%Y-%m-%dT%H:%M:%S.%3NZ') --name adgardhome-replica -t ghcr.io/bakito/adguardhome-sync:${AGH_SYNC_VERSION} .

kind-create:
	kind delete cluster
	kind create  cluster

kind-test:
	kubectl create namespace agh
	kubectl create configmap origin-conf -n agh --from-file testdata/e2e/AdGuardHome.yaml
	kubectl apply -n agh -f testdata/e2e/agh
	kubectl wait -n agh --for condition=Ready pod/adguardhome-origin --timeout=30s
	kubectl wait -n agh --for condition=Ready pod/adguardhome-replica --timeout=30s

	kubectl create configmap sync-conf -n agh --from-env-file=testdata/e2e/sync-conf.properties
	kubectl apply -n agh -f testdata/e2e/job-adguardhome-sync.yaml
	kubectl wait -n agh --for=condition=complete job/adguardhome-sync
