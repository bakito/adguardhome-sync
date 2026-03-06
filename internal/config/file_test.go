package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func Test_configFilePath(t *testing.T) {
	t.Run("should return the same value", func(t *testing.T) {
		path := uuid.NewString()
		result, err := configFilePath(path)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != path {
			t.Errorf("expected %v but got %v", path, result)
		}
	})
	t.Run("should the file in HOME dir", func(t *testing.T) {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		result, err := configFilePath("")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expected := filepath.Join(home, ".adguardhome-sync.yaml")
		if result != expected {
			t.Errorf("expected %v but got %v", expected, result)
		}
	})
}
