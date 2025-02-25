package config

import (
	"bytes"
	_ "embed"
	"os"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

//go:embed print-config.md
var printConfigTemplate string

func (ac *AppConfig) Print() error {
	out, err := ac.print(os.Environ())
	if err != nil {
		return err
	}

	logger.Infof(
		"Printing adguardhome-sync aggregated config (THE APPLICATION WILL NOT START IN THIS MODE):\n%s",
		out,
	)

	return nil
}

func (ac *AppConfig) print(env []string) (string, error) {
	config, err := yaml.Marshal(ac.Get())
	if err != nil {
		logger.Error(err)
		return "", err
	}

	t, err := template.New("printConfigTemplate").Parse(printConfigTemplate)
	if err != nil {
		return "", err
	}

	sort.Strings(env)

	var buf bytes.Buffer

	if err = t.Execute(&buf, map[string]interface{}{
		"AggregatedConfig":     string(config),
		"ConfigFilePath":       ac.filePath,
		"ConfigFileContent":    ac.content,
		"EnvironmentVariables": strings.Join(env, "\n"),
	}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
