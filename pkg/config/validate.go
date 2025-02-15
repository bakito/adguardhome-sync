package config

import (
	_ "embed"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

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
	schema := jsonschema.MustCompileString("adguardhome-sync/config", schemaData)

	// validateSchema
	return schema.Validate(yamlData)
}
