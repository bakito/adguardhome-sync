package config

import (
    "github.com/santhosh-tekuri/jsonschema/v5"
    "gopkg.in/yaml.v3"
    "io/ioutil"
    "log"
)

func main() {
    // Load YAML file
    yamlContent, err := ioutil.ReadFile("example.yaml")
    if err != nil {
        log.Fatalf("Error reading YAML file: %v", err)
    }

    // Convert YAML to JSON
    var yamlData interface{}
    err = yaml.Unmarshal(yamlContent, &yamlData)
    if err != nil {
        log.Fatalf("Error unmarshalling YAML: %v", err)
    }

    // Load JSON schema
    compiler := jsonschema.NewCompiler()
    schema, err := compiler.Compile("schema.json")
    if err != nil {
        log.Fatalf("Error compiling schema: %v", err)
    }

    // Validate
    if err := schema.Validate(yamlData); err != nil {
        log.Fatalf("Validation failed: %v", err)
    }

    fmt.Println("Validation succeeded!")
}
