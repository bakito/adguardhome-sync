package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func Test_buildCombinedTag(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		envTag   string
		expected string
	}{
		{"Both empty", "", "", ""},
		{"Prefix only", "PREFIX", "", "PREFIX"},
		{"Tag only", "", "TAG", "TAG"},
		{"Both set", "PREFIX", "TAG", "PREFIX_TAG"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildCombinedTag(tt.prefix, tt.envTag); got != tt.expected {
				t.Errorf("buildCombinedTag() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func Test_updateDocumentationSection(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		startMarker string
		endMarker   string
		newContent  string
		expected    string
	}{
		{
			"Standard update",
			"Before\n<!-- start -->\nOld\n<!-- end -->\nAfter",
			"<!-- start -->",
			"<!-- end -->",
			"New\n",
			"Before\n<!-- start -->\nNew\n<!-- end -->\nAfter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateDocumentationSection(
				tt.fileContent,
				tt.startMarker,
				tt.endMarker,
				tt.newContent,
			); got != tt.expected {
				t.Errorf("updateDocumentationSection() = %v, want %v", got, tt.expected)
			}
		})
	}
}

type innerStruct struct {
	Inner string `documentation:"Doc Inner" env:"INNER" yaml:"inner"`
}

type testStructEnv struct {
	Field1 string `documentation:"Doc 1"      env:"FIELD1"`
	Field2 int    `documentation:"Doc 2"      env:"FIELD2"`
	Nested innerStruct
	Origin string `documentation:"Doc Origin"` // Special case in code
}

func Test_writeEnvDocumentation(t *testing.T) {
	var buf bytes.Buffer
	writeEnvDocumentation(&buf, reflect.TypeOf(testStructEnv{}), "PRE")
	got := buf.String()

	expectedSubstrings := []string{
		"| PRE_FIELD1 (string) | string | Doc 1 |",
		"| PRE_FIELD2 (int) | int | Doc 2 |",
		"| PRE_INNER (string) | string | Doc Inner |",
		"| PRE_ORIGIN (string) | string | Doc Origin |",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(got, s) {
			t.Errorf("writeEnvDocumentation() output missing substring: %v\nGot:\n%v", s, got)
		}
	}
}

type testStructYAML struct {
	Field1   string `documentation:"Doc 1"        yaml:"field1"`
	Nested   innerStruct
	Replicas []string `documentation:"Doc Replicas" yaml:"replicas"`
}

func Test_writeYAMLDocumentation(t *testing.T) {
	var buf bytes.Buffer
	writeYAMLDocumentation(&buf, reflect.TypeOf(testStructYAML{}), "", "")
	got := buf.String()

	expectedSubstrings := []string{
		"field1: # (string) Doc 1",
		"inner: # (string) Doc Inner",
		"replicas: # (string) Doc Replicas",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(got, s) {
			t.Errorf("writeYAMLDocumentation() output missing substring: %v\nGot:\n%v", s, got)
		}
	}
}
