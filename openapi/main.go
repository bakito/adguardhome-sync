package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
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

	if comp, ok := schema["components"]; ok {
		if rb, ok := comp.(map[string]interface{})["requestBodies"]; ok {
			for k, v := range rb.(map[string]interface{}) {
				v.(map[string]interface{})["x-go-name"] = k + "Body"
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
