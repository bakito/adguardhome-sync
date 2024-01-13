package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/bakito/adguardhome-sync/pkg/utils"
	"github.com/caarlos0/env/v10"
	"github.com/spf13/cobra"
)

var (
	envReplicasURLPattern = regexp.MustCompile(`^REPLICA(\d+)_URL=(.*)`)
	logger                = log.GetLogger("config")
)

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

func Get(configFile string, cmd *cobra.Command) (*types.Config, error) {
	path := configFilePath(configFile)

	cfg := &types.Config{
		Replica: &types.AdGuardInstance{},
	}

	// read yaml config
	if err := readFile(cfg, path); err != nil {
		return nil, err
	}

	// overwrite from command flags
	if err := readFlags(cfg, cmd); err != nil {
		return nil, err
	}

	// overwrite from env vars

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	if err := env.ParseWithOptions(&cfg.Origin, env.Options{Prefix: "ORIGIN_"}); err != nil {
		return nil, err
	}
	if err := env.ParseWithOptions(cfg.Replica, env.Options{Prefix: "REPLICA_"}); err != nil {
		return nil, err
	}

	if cfg.Replica != nil &&
		cfg.Replica.URL == "" &&
		cfg.Replica.Username == "" {
		cfg.Replica = nil
	}

	if len(cfg.Replicas) > 0 && cfg.Replica != nil {
		return nil, errors.New("mixed replica config in use. " +
			"Do not use single replica and numbered (list) replica config combined")
	}

	handleDeprecatedEnvVars(cfg)

	if cfg.Replica != nil {
		cfg.Replicas = []types.AdGuardInstance{*cfg.Replica}
		cfg.Replica = nil
	}

	cfg.Replicas = enrichReplicasFromEnv(cfg.Replicas)

	return cfg, nil
}

func handleDeprecatedEnvVars(cfg *types.Config) {
	if val, ok := checkDeprecatedEnvVar("RUNONSTART", "RUN_ON_START"); ok {
		cfg.RunOnStart, _ = strconv.ParseBool(val)
	}
	if val, ok := checkDeprecatedEnvVar("API_DARKMODE", "API_DARK_MODE"); ok {
		cfg.API.DarkMode, _ = strconv.ParseBool(val)
	}
	if val, ok := checkDeprecatedEnvVar("FEATURES_GENERALSETTINGS", "FEATURES_GENERAL_SETTINGS"); ok {
		cfg.Features.GeneralSettings, _ = strconv.ParseBool(val)
	}
	if val, ok := checkDeprecatedEnvVar("FEATURES_QUERYLOGCONFIG", "FEATURES_QUERY_LOG_CONFIG"); ok {
		cfg.Features.QueryLogConfig, _ = strconv.ParseBool(val)
	}
	if val, ok := checkDeprecatedEnvVar("FEATURES_STATSCONFIG", "FEATURES_STATS_CONFIG"); ok {
		cfg.Features.StatsConfig, _ = strconv.ParseBool(val)
	}
	if val, ok := checkDeprecatedEnvVar("FEATURES_CLIENTSETTINGS", "FEATURES_CLIENT_SETTINGS"); ok {
		cfg.Features.ClientSettings, _ = strconv.ParseBool(val)
	}
	if val, ok := checkDeprecatedEnvVar("FEATURES_DHCP_SERVERCONFIG", "FEATURES_DHCP_SERVER_CONFIG"); ok {
		cfg.Features.DHCP.ServerConfig, _ = strconv.ParseBool(val)
	}
	if val, ok := checkDeprecatedEnvVar("FEATURES_DHCP_STATICLEASES", "FEATURES_DHCP_STATIC_LEASES"); ok {
		cfg.Features.DHCP.StaticLeases, _ = strconv.ParseBool(val)
	}
	if val, ok := checkDeprecatedEnvVar("FEATURES_DNS_ACCESSLISTS", "FEATURES_DNS_ACCESS_LISTS"); ok {
		cfg.Features.DNS.AccessLists, _ = strconv.ParseBool(val)
	}
	if val, ok := checkDeprecatedEnvVar("FEATURES_DNS_SERVERCONFIG", "FEATURES_DNS_SERVER_CONFIG"); ok {
		cfg.Features.DNS.ServerConfig, _ = strconv.ParseBool(val)
	}

	if cfg.Replica != nil {
		if val, ok := checkDeprecatedEnvVar("REPLICA_WEBURL", "REPLICA_WEB_URL"); ok {
			cfg.Replica.WebURL = val
		}
		if val, ok := checkDeprecatedEnvVar("REPLICA_AUTOSETUP", "REPLICA_AUTO_SETUP"); ok {
			cfg.Replica.AutoSetup, _ = strconv.ParseBool(val)
		}
		if val, ok := checkDeprecatedEnvVar("REPLICA_INTERFACENAME", "REPLICA_INTERFACE_NAME"); ok {
			cfg.Replica.InterfaceName = val
		}
		if val, ok := checkDeprecatedEnvVar("REPLICA_DHCPSERVERENABLED", "REPLICA_DHCP_SERVER_ENABLED"); ok {
			if b, err := strconv.ParseBool(val); err != nil {
				cfg.Replica.DHCPServerEnabled = utils.Ptr(b)
			}
		}
	}
}

func checkDeprecatedEnvVar(oldName string, newName string) (string, bool) {
	old, oldOK := os.LookupEnv(oldName)
	if oldOK {
		logger.With("deprecated", oldName, "replacement", newName).
			Warn("Deprecated env variable is used, please use the correct one")
	}
	new, newOK := os.LookupEnv(newName)
	if newOK {
		return new, true
	}
	return old, oldOK
}

func checkDeprecatedReplicaEnvVar(oldPattern string, newPattern string, replicaID int) (string, bool) {
	return checkDeprecatedEnvVar(fmt.Sprintf(oldPattern, replicaID), fmt.Sprintf(newPattern, replicaID))
}

// Manually collect replicas from env.
func enrichReplicasFromEnv(initialReplicas []types.AdGuardInstance) []types.AdGuardInstance {
	var replicas []types.AdGuardInstance
	for _, v := range os.Environ() {
		if envReplicasURLPattern.MatchString(v) {
			sm := envReplicasURLPattern.FindStringSubmatch(v)
			index, _ := strconv.Atoi(sm[1])
			if index > len(initialReplicas) {
				replicas = append(replicas, types.AdGuardInstance{URL: sm[2]})
			} else {
				re := initialReplicas[index-1]
				re.URL = sm[2]
				replicas = append(replicas, re)
			}
		}
	}

	if len(replicas) == 0 {
		replicas = initialReplicas
	}

	for i := range replicas {
		reID := i + 1

		if val, ok := os.LookupEnv(fmt.Sprintf("REPLICA%d_WEB_URL", reID)); ok {
			replicas[i].WebURL = val
		}
		if val, ok := os.LookupEnv(fmt.Sprintf("REPLICA%d_USERNAME", reID)); ok {
			replicas[i].Username = val
		}
		if val, ok := os.LookupEnv(fmt.Sprintf("REPLICA%d_PASSWORD", reID)); ok {
			replicas[i].Password = val
		}
		if val, ok := os.LookupEnv(fmt.Sprintf("REPLICA%d_COOKIE", reID)); ok {
			replicas[i].Cookie = val
		}
		if val, ok := checkDeprecatedReplicaEnvVar("REPLICA%d_APIPATH", "REPLICA%d_API_PATH", reID); ok {
			replicas[i].APIPath = val
		}
		if val, ok := checkDeprecatedReplicaEnvVar("REPLICA%d_INSECURESKIPVERIFY", "REPLICA%d_INSECURE_SKIP_VERIFY", reID); ok {
			replicas[i].InsecureSkipVerify = strings.EqualFold(val, "true")
		}
		if val, ok := checkDeprecatedReplicaEnvVar("REPLICA%d_AUTOSETUP", "REPLICA%d_AUTO_SETUP", reID); ok {
			replicas[i].AutoSetup = strings.EqualFold(val, "true")
		}
		if val, ok := checkDeprecatedReplicaEnvVar("REPLICA%d_INTERFACENAME", "REPLICA%d_INTERFACE_NAME", reID); ok {
			replicas[i].InterfaceName = val
		}

		if dhcpEnabled, ok := checkDeprecatedReplicaEnvVar("REPLICA%d_DHCPSERVERENABLED", "REPLICA%d_DHCP_SERVER_ENABLED", reID); ok {
			if strings.EqualFold(dhcpEnabled, "true") {
				replicas[i].DHCPServerEnabled = utils.Ptr(true)
			} else if strings.EqualFold(dhcpEnabled, "false") {
				replicas[i].DHCPServerEnabled = utils.Ptr(false)
			}
		}
		if replicas[i].APIPath == "" {
			replicas[i].APIPath = "/control"
		}
	}

	return replicas
}
