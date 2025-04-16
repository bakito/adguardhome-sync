package config

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

const schemaURL = "config-schema.json"

//go:embed config-schema.json
var schemaData string

func validateSchema(cfgFile string) error {
	// ignore if file not exists
	if _, err := os.Stat(cfgFile); err != nil {
		// Config file does not exist or is not readable - ignore it
		//nolint:nilerr
		return nil
	}
	// Load YAML file
	yamlContent, err := os.ReadFile(cfgFile)
	if err != nil {
		return fmt.Errorf("config file %q is invalid: %w", cfgFile, err)
	}

	return validateYAML(yamlContent)
}

func validateYAML(yamlContent []byte) error {
	if yamlContent == nil || strings.TrimSpace(string(yamlContent)) == "" {
		return nil
	}

	// Convert YAML to JSON
	var yamlData any
	err := yaml.Unmarshal(yamlContent, &yamlData)
	if err != nil {
		return err
	}

	// Load JSON schema
	sch, err := jsonschema.UnmarshalJSON(strings.NewReader(schemaData))
	if err != nil {
		return err
	}

	c := jsonschema.NewCompiler()
	if err := c.AddResource(schemaURL, sch); err != nil {
		return err
	}
	schema := c.MustCompile(schemaURL)
	// validateSchema
	return schema.Validate(yamlData)
}
