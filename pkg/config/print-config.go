package config

import (
	"bytes"
	_ "embed"
	"os"
	"runtime"
	"sort"
	"strings"
	"text/template"

	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/bakito/adguardhome-sync/version"
	"gopkg.in/yaml.v3"
)

//go:embed print-config.md
var printConfigTemplate string

func (ac *AppConfig) Print() error {
	originVersion := aghVersion(ac.cfg.Origin)
	var replicaVersions []string
	for _, replica := range ac.cfg.Replicas {
		replicaVersions = append(replicaVersions, aghVersion(replica))
	}

	out, err := ac.print(os.Environ(), originVersion, replicaVersions)
	if err != nil {
		return err
	}

	logger.Infof(
		"Printing adguardhome-sync aggregated config (THE APPLICATION WILL NOT START IN THIS MODE):\n%s",
		out,
	)

	return nil
}

func aghVersion(i types.AdGuardInstance) string {
	cl, err := client.New(i)
	if err != nil {
		return "N/A"
	}
	stats, err := cl.Status()
	if err != nil {
		return "N/A"
	}
	return stats.Version
}

func (ac *AppConfig) print(env []string, originVersion string, replicaVersions []string) (string, error) {
	config, err := yaml.Marshal(ac.Get())
	if err != nil {
		return "", err
	}

	funcMap := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"inc": func(i int) int {
			return i + 1
		},
	}

	t, err := template.New("printConfigTemplate").Funcs(funcMap).Parse(printConfigTemplate)
	if err != nil {
		return "", err
	}

	sort.Strings(env)

	var buf bytes.Buffer

	if err = t.Execute(&buf, map[string]interface{}{
		"Version":              version.Version,
		"Build":                version.Build,
		"OperatingSystem":      runtime.GOOS,
		"Architecture":         runtime.GOARCH,
		"AggregatedConfig":     string(config),
		"ConfigFilePath":       ac.filePath,
		"ConfigFileContent":    ac.content,
		"EnvironmentVariables": strings.Join(env, "\n"),
		"OriginVersion":        originVersion,
		"ReplicaVersions":      replicaVersions,
	}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
