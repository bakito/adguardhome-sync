package config

import (
	_ "embed"
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
		return nil
	}
	// Load YAML file
	yamlContent, err := os.ReadFile(cfgFile)
	if err != nil {
		return err
	}

	return validateYAML(yamlContent)
}

func validateYAML(yamlContent []byte) error {
	// Convert YAML to JSON
	var yamlData interface{}
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
