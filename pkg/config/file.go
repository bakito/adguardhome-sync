package config

import (
	"fmt"
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

func configFilePath(configFile string) string {
	if configFile == "" {
		// Find home directory.
		home, err := os.UserConfigDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return filepath.Join(home, ".adguardhome-sync")
	}
	return configFile
}
