# Run go fmt against code
fmt:
	golangci-lint run --fix

# Run go mod tidy
tidy:
	go mod tidy

# Run tests
test: test-ci fmt

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

start-replica:
	podman run --pull always --rm -it -p 9090:80 -p 9091:3000  adguard/adguardhome