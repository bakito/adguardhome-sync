// Print the available environment variables
package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/bakito/adguardhome-sync/internal/types"
)

const (
	envStartMarker  = "<!-- env-doc-start -->"
	envEndMarker    = "<!-- env-doc-end -->"
	yamlStartMarker = "<!-- yaml-doc-start -->"
	yamlEndMarker   = "<!-- yaml-doc-end -->"
)

func main() {
	slog.Info("Reading README.md")
	content, err := os.ReadFile("README.md")
	if err != nil {
		slog.Error("Error reading README.md", "error", err)
		os.Exit(1)
	}

	fileContent := string(content)

	slog.Info("Generating environment variables")
	fileContent = generateEnvDocumentation(fileContent)

	slog.Info("Generating yaml configuration")
	fileContent = generateYAMLDocumentation(fileContent)

	slog.Info("Writing README.md")
	err = os.WriteFile("README.md", []byte(fileContent), 0o644)
	if err != nil {
		slog.Error("Error writing README.md", "error", err)
		os.Exit(1)
	}
}

func generateEnvDocumentation(fileContent string) string {
	var buf strings.Builder
	_, _ = buf.WriteString("| Name | Type | Description |\n")
	_, _ = buf.WriteString("| :--- | ---- |:----------- |\n")
	writeEnvDocumentation(&buf, reflect.TypeOf(types.Config{}), "")

	return updateDocumentationSection(fileContent, envStartMarker, envEndMarker, buf.String())
}

func generateYAMLDocumentation(fileContent string) string {
	var buf strings.Builder
	_, _ = buf.WriteString("```yaml\n")
	writeYAMLDocumentation(&buf, reflect.TypeOf(types.Config{}), "", "")
	_, _ = buf.WriteString("```\n")

	return updateDocumentationSection(fileContent, yamlStartMarker, yamlEndMarker, buf.String())
}

func updateDocumentationSection(fileContent, startMarker, endMarker, newContent string) string {
	startIdx := strings.Index(fileContent, startMarker)
	endIdx := strings.Index(fileContent, endMarker)

	if startIdx == -1 || endIdx == -1 {
		slog.Error(fmt.Sprintf("Could not find markers %s and %s in README.md", startMarker, endMarker))
		os.Exit(1)
	}

	return fileContent[:startIdx+len(startMarker)] + "\n" + newContent + fileContent[endIdx:]
}

func writeEnvDocumentation(w io.Writer, t reflect.Type, prefix string) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	for _, field := range reflect.VisibleFields(t) {
		if field.PkgPath != "" {
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

		combinedTag := buildCombinedTag(prefix, envTag)

		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		if ft.Kind() == reflect.Struct && ft.Name() != "Time" {
			writeEnvDocumentation(w, ft, strings.TrimSuffix(combinedTag, "_"))
		} else if envTag != "" {
			envVar := strings.Trim(combinedTag, "_") + " (" + ft.Kind().String() + ")"
			docs := field.Tag.Get("documentation")
			_, _ = fmt.Fprintf(w, "| %s | %s | %s |\n", envVar, ft.Kind().String(), docs)
		}
	}
}

func writeYAMLDocumentation(w io.Writer, t reflect.Type, firstPrefix, otherPrefix string) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	var i int
	for _, field := range reflect.VisibleFields(t) {
		if field.PkgPath != "" {
			continue
		}

		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "-" {
			continue
		}
		yamlTag = strings.TrimSuffix(yamlTag, ",omitempty")

		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		pf := otherPrefix
		if i == 0 {
			pf = firstPrefix
		}

		newFirstPrefix := pf + "  "
		newOtherPrefix := otherPrefix + "  "

		if yamlTag == "replicas" && ft.Kind() == reflect.Slice {
			ft = ft.Elem()
			newFirstPrefix += "- "
			newOtherPrefix += "  "
		}

		if yamlTag != "" {
			docs := field.Tag.Get("documentation")
			_, _ = fmt.Fprintf(w, "%s%s: # (%s) %s\n", pf, yamlTag, ft.Kind().String(), docs)
			i++
		}

		if ft.Kind() == reflect.Struct && ft.Name() != "Time" {
			writeYAMLDocumentation(w, ft, newFirstPrefix, newOtherPrefix)
		}
	}
}

func buildCombinedTag(prefix, envTag string) string {
	if prefix != "" && envTag != "" {
		return prefix + "_" + envTag
	} else if prefix != "" {
		return prefix
	}
	return envTag
}
