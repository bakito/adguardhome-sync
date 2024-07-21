package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/bakito/adguardhome-sync/pkg/utils"
	"github.com/caarlos0/env/v11"
)

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
func enrichReplicasFromEnv(initialReplicas []types.AdGuardInstance) ([]types.AdGuardInstance, error) {
	var replicas []types.AdGuardInstance
	for _, v := range os.Environ() {
		if envReplicasURLPattern.MatchString(v) {
			sm := envReplicasURLPattern.FindStringSubmatch(v)
			id, _ := strconv.Atoi(sm[1])

			if id <= 0 {
				return nil, fmt.Errorf("numbered replica env variables must have a number id >= 1, got %q", v)
			}

			if id > len(initialReplicas) {
				replicas = append(replicas, types.AdGuardInstance{URL: sm[2]})
			} else {
				re := initialReplicas[id-1]
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

		// keep previously set value
		replicaDhcpServer := replicas[i].DHCPServerEnabled
		replicas[i].DHCPServerEnabled = nil
		if err := env.ParseWithOptions(&replicas[i], env.Options{Prefix: fmt.Sprintf("REPLICA%d_", reID)}); err != nil {
			return nil, err
		}
		if replicas[i].DHCPServerEnabled == nil {
			replicas[i].DHCPServerEnabled = replicaDhcpServer
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

	return replicas, nil
}
