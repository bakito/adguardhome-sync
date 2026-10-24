package main

import (
	"reflect"

	"github.com/bakito/adguardhome-sync/internal/metrics"
	"github.com/bakito/adguardhome-sync/internal/types"
	"github.com/bakito/docs-gen/docs"
	"github.com/bakito/docs-gen/pkg/env"
	"github.com/bakito/docs-gen/pkg/template"
	"github.com/bakito/docs-gen/pkg/yaml"
)

const (
	envStartMarker     = "<!-- env-doc-start -->"
	envEndMarker       = "<!-- env-doc-end -->"
	yamlStartMarker    = "<!-- yaml-doc-start -->"
	yamlEndMarker      = "<!-- yaml-doc-end -->"
	metricsStartMarker = "<!-- metrics-doc-start -->"
	metricsEndMarker   = "<!-- metrics-doc-end -->"
)

func main() {
	docs.UpdateDocumentation("README.md",
		env.UpdateDocumentationWithCustomizer[types.Config](envStartMarker, envEndMarker, envTagCustomizer),
		yaml.UpdateDocumentationWithCustomizer[types.Config](yamlStartMarker, yamlEndMarker, yamlPrefixCustomizer),
		template.UpdateDocumentation(metrics.MetricsDoc(), metricsStartMarker, metricsEndMarker,
			"| {{ .Key }} | {{ .Value }} |\n",
			template.WithPrefix("| Metric | Description |\n| ------ | ----------- |\n"),
		),
	)
}

func yamlPrefixCustomizer(yamlTag string, prefix *yaml.Prefix) {
	if yamlTag == "replicas" && prefix.FieldType.Kind() == reflect.Slice {
		prefix.FieldType = prefix.FieldType.Elem()
		prefix.First += "- "
		prefix.Other += "  "
	}
}

func envTagCustomizer(envTag string, field reflect.StructField) string {
	if envTag == "" {
		switch field.Name {
		case "Origin":
			envTag = "ORIGIN"
		case "Replica":
			envTag = "REPLICA#"
		}
	}
	return envTag
}
