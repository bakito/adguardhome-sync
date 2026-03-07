package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestConfigFilePathWithSpecifiedPath(t *testing.T) {
	path := uuid.NewString()
	result, err := configFilePath(path)
	if err != nil {
		t.Fatalf("configFilePath(%s) error = %v, want nil", path, err)
	}
	if result != path {
		t.Errorf("configFilePath(%s) result = %s, want %s", path, result, path)
	}
}

func TestConfigFilePathWithEmptyPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to os.UserHomeDir(): %v", err)
	}
	result, err := configFilePath("")
	if err != nil {
		t.Fatalf("configFilePath('') error = %v, want nil", err)
	}
	expected := filepath.Join(home, ".adguardhome-sync.yaml")
	if result != expected {
		t.Errorf("configFilePath('') result = %s, want %s", result, expected)
	}
}
