package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

type replicaConfig struct {
	Versions []string `yaml:"versions"`
}

type syncConfig struct {
	Enabled bool `yaml:"enabled"`
}

type testValues struct {
	Replica replicaConfig `yaml:"replica"`
	Mode    string        `yaml:"mode"`
	Sync    syncConfig    `yaml:"sync"`
}

func Test_updateReplicaVersions(t *testing.T) {
	initialYAML := `replica:
  versions:
    - v0.107.0
    - v0.107.1
    - latest
mode: env
sync:
  enabled: true
`
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "values.yaml")
	err := os.WriteFile(testFile, []byte(initialYAML), 0o600)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err = updateReplicaVersions(testFile, "v0.107.68", "v0.107.79", "latest")
	if err != nil {
		t.Fatalf("updateReplicaVersions failed: %v", err)
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	var parsed testValues
	err = yaml.Unmarshal(data, &parsed)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	expectedVersions := []string{"v0.107.68", "v0.107.79", "latest"}
	if !reflect.DeepEqual(parsed.Replica.Versions, expectedVersions) {
		t.Errorf("Versions = %v, want %v", parsed.Replica.Versions, expectedVersions)
	}
	if parsed.Mode != "env" {
		t.Errorf("Mode = %q, want %q", parsed.Mode, "env")
	}
	if !parsed.Sync.Enabled {
		t.Error("expected Sync.Enabled to be true")
	}
}
