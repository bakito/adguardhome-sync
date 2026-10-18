package config

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/bakito/adguardhome-sync/internal/types"
)

func FuzzYAMLConfigParsing(f *testing.F) {
	// Seed corpus with valid and malformed YAML configs
	seeds := []string{
		"",
		"{}",
		`cron: "*/5 * * * *"`,
		`origin:
  url: "https://localhost:3000"
  username: "admin"
  password: "secret"`,
		`replica:
  url: "https://replica:3000"`,
		`replicas:
  - url: "https://r1:3000"
  - url: "https://r2:3000"`,
		`api:
  port: 8080
  username: "apiuser"
  password: "apipass"`,
		`features:
  dns:
    accessLists: true
    serverConfig: true`,
		// Malformed YAML
		`{`,
		`origin:
  url:`,
		`- - -`,
		`!!str 123`,
		`api:
  port: notanumber`,
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(_ *testing.T, yamlInput string) {
		cfg := &types.Config{}
		// Unmarshal may fail for invalid YAML; that's acceptable.
		_ = yaml.Unmarshal([]byte(yamlInput), cfg)
	})
}
