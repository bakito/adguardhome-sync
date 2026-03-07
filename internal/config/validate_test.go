package config

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"gopkg.in/yaml.v3"

	"github.com/bakito/adguardhome-sync/internal/types"
)

func TestValidateSchema(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		expectFail bool
	}{
		{"Should be valid", "../../testdata/config/config-valid.yaml", false},
		{"Should be valid if file doesn't exist", "../../testdata/config/foo.bar", false},
		{"Should fail if file is not yaml", "../../go.mod", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchema(tt.configFile)
			if (err != nil) != tt.expectFail {
				t.Errorf("validateSchema(%s) error = %v, expectFail %v", tt.configFile, err, tt.expectFail)
			}
		})
	}
}

func TestValidateConfigWithAllFieldsRandomlyPopulated(t *testing.T) {
	cfg := &types.Config{}

	err := faker.FakeData(cfg)
	if err != nil {
		t.Fatalf("failed to faker.FakeData(cfg): %v", err)
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to yaml.Marshal(&cfg): %v", err)
	}

	err = validateYAML(data)
	if err != nil {
		t.Errorf("validateYAML(data) error = %v, want nil", err)
	}
}

func TestValidateConfigWithEmptyFile(t *testing.T) {
	var data []byte
	err := validateYAML(data)
	if err != nil {
		t.Errorf("validateYAML(data) error = %v, want nil", err)
	}
}
