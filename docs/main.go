// Print the available environment variables
package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/bakito/adguardhome-sync/pkg/types"
)

func main() {
	_, _ = fmt.Println("| Name | Type | Description |")
	_, _ = fmt.Println("| :--- | ---- |:----------- |")
	printEnvTags(reflect.TypeOf(types.Config{}), "")
}

// printEnvTags recursively prints all fields with `env` tags.
func printEnvTags(t reflect.Type, prefix string) {
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
			default:
				continue
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
			printEnvTags(ft, strings.TrimSuffix(combinedTag, "_"))
		} else if envTag != "" {
			envVar := strings.Trim(combinedTag, "_") + " (" + ft.Kind().String() + ")"
			docs := field.Tag.Get("documentation")

			_, _ = fmt.Printf("| %s | %s | %s |\n", envVar, ft.Kind().String(), docs)
		}
	}
}
