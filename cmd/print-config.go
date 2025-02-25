package cmd

import (
	"bytes"
	_ "embed"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/bakito/adguardhome-sync/pkg/types"
	"gopkg.in/yaml.v3"
)

//go:embed print-config.md
var printConfigTemplate string

func printConfig(cfg *types.Config, usedCfgFile string, cfgContent string) error {
	config, err := yaml.Marshal(cfg)
	if err != nil {
		logger.Error(err)
		return err
	}

	t, err := template.New("printConfigTemplate").Parse(printConfigTemplate)
	if err != nil {
		return err
	}

	env := os.Environ()
	sort.Strings(env)

	var buf bytes.Buffer

	if err = t.Execute(&buf, map[string]interface{}{
		"AggregatedConfig":     string(config),
		"ConfigFilePath":       usedCfgFile,
		"ConfigFileContent":    cfgContent,
		"EnvironmentVariables": strings.Join(env, "\n"),
	}); err != nil {
		return err
	}

	logger.Infof(
		"Printing adguardhome-sync aggregated config (THE APPLICATION WILL NOT START IN THIS MODE):\n%s",
		buf.String(),
	)

	return nil
}
