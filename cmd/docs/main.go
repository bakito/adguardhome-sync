package main

import (
	"github.com/bakito/adguardhome-sync/internal/types"
	"github.com/bakito/docs-gen/docs"
	"github.com/bakito/docs-gen/pkg/env"
	"github.com/bakito/docs-gen/pkg/yaml"
)

const (
	envStartMarker  = "<!-- env-doc-start -->"
	envEndMarker    = "<!-- env-doc-end -->"
	yamlStartMarker = "<!-- yaml-doc-start -->"
	yamlEndMarker   = "<!-- yaml-doc-end -->"
)

func main() {
	docs.UpdateDocumentation("README.md",
		env.UpdateDocumentation[types.Config](envStartMarker, envEndMarker),
		yaml.UpdateDocumentation[types.Config](yamlStartMarker, yamlEndMarker),
	)
}
