// Print the available environment variables
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/bakito/adguardhome-sync/pkg/types"
)

func main() {
	// Read the README.md file
	content, err := os.ReadFile("README.md")
	if err != nil {
		log.Fatal(err)
	}

	// Convert to string for easier manipulation
	fileContent := string(content)

	// Generate the environment variables documentation
	var buf strings.Builder
	_, _ = buf.WriteString("| Name | Type | Description |\n")
	_, _ = buf.WriteString("| :--- | ---- |:----------- |\n")
	printEnvTags(&buf, reflect.TypeOf(types.Config{}), "")

	// Find the markers and replace content between them
	startMarker := "<!-- env-doc-start -->"
	endMarker := "<!-- env-doc-end -->"

	start := strings.Index(fileContent, startMarker)
	end := strings.Index(fileContent, endMarker)

	if start == -1 || end == -1 {
		log.Fatal("Could not find markers in README.md")
	}

	// Construct new content
	newContent := fileContent[:start+len(startMarker)] + "\n" + buf.String() + fileContent[end:]

	// Write back to README.md
	err = os.WriteFile("README.md", []byte(newContent), 0o644)
	if err != nil {
		log.Fatal(err)
	}
}

// printEnvTags recursively prints all fields with `env` tags.
func printEnvTags(w io.Writer, t reflect.Type, prefix string) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	for _, field := range reflect.VisibleFields(t) {
		if field.PkgPath != "" { // unexported field
			continue
		}

		envTag := field.Tag.Get("env")
		if envTag == "" {
			switch field.Name {
			case "Origin":
				envTag = "ORIGIN"
			case "Replica":
				envTag = "REPLICA#"
			}
		}
		combinedTag := envTag
		if prefix != "" && envTag != "" {
			combinedTag = prefix + "_" + envTag
		} else if prefix != "" {
			combinedTag = prefix
		}

		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		if ft.Kind() == reflect.Struct && ft.Name() != "Time" { // skip time.Time
			printEnvTags(w, ft, strings.TrimSuffix(combinedTag, "_"))
		} else if envTag != "" {
			envVar := strings.Trim(combinedTag, "_") + " (" + ft.Kind().String() + ")"
			docs := field.Tag.Get("documentation")

			_, _ = fmt.Fprintf(w, "| %s | %s | %s |\n", envVar, ft.Kind().String(), docs)
		}
	}
}
