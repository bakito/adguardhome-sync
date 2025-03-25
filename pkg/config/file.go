package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/bakito/adguardhome-sync/pkg/types"
)

func readFile(cfg *types.Config, path string) (string, error) {
	var content string
	if _, err := os.Stat(path); err == nil {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		content = string(b)
		if err := yaml.Unmarshal(b, cfg); err != nil {
			return "", err
		}
	}
	return content, nil
}

func configFilePath(configFile string) (string, error) {
	if configFile == "" {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".adguardhome-sync.yaml"), nil
	}
	return configFile, nil
}
