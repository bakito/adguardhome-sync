package main

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_addFakeTags(t *testing.T) {
	schema := map[string]any{
		"components": map[string]any{
			"schemas": map[string]any{
				"Stats": map[string]any{
					"properties": map[string]any{
						"blocked_filtering":     map[string]any{},
						"dns_queries":           map[string]any{},
						"replaced_parental":     map[string]any{},
						"replaced_safebrowsing": map[string]any{},
					},
				},
			},
		},
	}

	addFakeTags(schema)

	properties := []string{
		"blocked_filtering",
		"dns_queries",
		"replaced_parental",
		"replaced_safebrowsing",
	}

	for _, prop := range properties {
		t.Run(prop, func(t *testing.T) {
			if val, ok, _ := unstructured.NestedString(
				schema,
				"components",
				"schemas",
				"Stats",
				"properties",
				prop,
				"x-oapi-codegen-extra-tags",
				"faker",
			); !ok ||
				val != "slice_len=24" {
				t.Errorf("addFakeTags() did not set expected faker tag for property %s, got %v", prop, val)
			}
		})
	}
}

func Test_correctEntries(_ *testing.T) {
	// Currently empty, but let's test that it doesn't panic
	correctEntries(make(map[string]any))
}
