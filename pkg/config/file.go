package config

import (
	"os"
	"path/filepath"

	"github.com/bakito/adguardhome-sync/pkg/types"
	"gopkg.in/yaml.v3"
)

func readFile(cfg *types.Config, path string) error {
	if _, err := os.Stat(path); err == nil {
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(b, cfg); err != nil {
			return err
		}
	}
	return nil
}

func configFilePath(configFile string) (string, error) {
	if configFile == "" {
		// Find home directory.
		home, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".adguardhome-sync"), nil
	}
	return configFile, nil
}
