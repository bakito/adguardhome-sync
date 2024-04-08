//go:build tools
// +build tools

package tools

import (
	_ "github.com/bakito/semver"
	_ "github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/onsi/ginkgo/v2/ginkgo"
	_ "go.uber.org/mock/mockgen"
	_ "k8s.io/code-generator/cmd/deepcopy-gen"
)
