# Run go fmt against code
fmt:
	go fmt ./...
	gofmt -s -w .

# Run go vet against code
vet:
	go vet ./...

# Run golangci-lint
lint:
	golangci-lint run

# Run go mod tidy
tidy:
	go mod tidy

# Run tests
test: mocks tidy fmt vet
	go test ./...  -coverprofile=coverage.out
	go tool cover -func=coverage.out

mocks: mockgen
	mockgen -destination pkg/mocks/client/mock.go github.com/bakito/adguardhome-sync/pkg/client Client

release: semver
	@version=$$(semver); \
	git tag -s $$version -m"Release $$version"
	goreleaser --rm-dist

test-release:
	goreleaser --skip-publish --snapshot --rm-dist

semver:
ifeq (, $(shell which semver))
 $(shell go get -u github.com/bakito/semver)
endif

mockgen:
ifeq (, $(shell which mockgen))
 $(shell go get github.com/golang/mock/mockgen@v1.5)
endif