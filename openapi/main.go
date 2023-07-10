package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func main() {
	version := "master"
	fileName := "schema-master.yaml"
	if len(os.Args) > 1 {
		version = os.Args[1]
		fileName = "schema.yaml"
	}
	log.Printf("Patching schema version %s\n", version)

	resp, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/AdguardTeam/AdGuardHome/%s/openapi/openapi.yaml", version))
	if err != nil {
		log.Fatalln(err)
	}
	defer func() { resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	schema := make(map[string]interface{})
	err = yaml.Unmarshal(data, &schema)
	if err != nil {
		log.Fatalln(err)
	}

	if requestBodies, ok, _ := unstructured.NestedMap(schema, "components", "requestBodies"); ok {
		for k := range requestBodies {
			_ = unstructured.SetNestedField(schema, k+"Body", "components", "requestBodies", k, "x-go-name")
		}
	}

	if dnsInfo, ok, _ := unstructured.NestedMap(schema,
		"paths", "/dns_info", "get", "responses", "200", "content", "application/json", "schema"); ok {
		if allOf, ok, _ := unstructured.NestedSlice(dnsInfo, "allOf"); ok && len(allOf) == 2 {
			delete(dnsInfo, "allOf")
			if err := unstructured.SetNestedMap(schema, allOf[0].(map[string]interface{}),
				"paths", "/dns_info", "get", "responses", "200", "content", "application/json", "schema"); err != nil {
				log.Fatalln(err)
			}
		}
	}
	b, err := yaml.Marshal(&schema)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Writing schema file tmp/%s", fileName)
	err = os.WriteFile("tmp/"+fileName, b, 0o600)
	if err != nil {
		log.Fatalln(err)
	}
}
