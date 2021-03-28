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
test: tidy fmt vet
	go test ./...  -coverprofile=coverage.out
	go tool cover -func=coverage.out

release:
	@version=$$(go run version/semver/main.go); \
	git tag -s $$version -m"Release $$version"
	goreleaser --rm-dist

test-release:
	goreleaser --skip-publish --snapshot --rm-dist