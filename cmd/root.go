package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/bakito/adguardhome-sync/version"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	configCron            = "cron"
	configRunOnStart      = "runOnStart"
	configPrintConfigOnly = "PRINT_CONFIG_ONLY"
	configContinueOnError = "CONTINUE_ON_ERROR"

	configAPIPort     = "api.port"
	configAPIUsername = "api.username"
	configAPIPassword = "api.password"
	configAPIDarkMode = "api.darkMode"

	configFeatureDHCPServerConfig = "features.dhcp.SERVER_CONFIG"
	configFeatureDHCPStaticLeases = "features.dhcp.STATIC_LEASES"
	configFeatureDNServerConfig   = "features.dns.SERVER_CONFIG"
	configFeatureDNSPAccessLists  = "features.dns.ACCESS_LISTS"
	configFeatureDNSRewrites      = "features.dns.rewrites"
	configFeatureGeneralSettings  = "features.GENERAL_SETTINGS"
	configFeatureQueryLogConfig   = "features.QUERY_LOG_CONFIG"
	configFeatureStatsConfig      = "features.STATS_CONFIG"
	configFeatureClientSettings   = "features.CLIENT_SETTINGS"
	configFeatureServices         = "features.services"
	configFeatureFilters          = "features.filters"

	configOriginURL                = "origin.url"
	configOriginWebURL             = "origin.webURL"
	configOriginAPIPath            = "origin.apiPath"
	configOriginUsername           = "origin.username"
	configOriginPassword           = "origin.password"
	configOriginCookie             = "origin.cookie"
	configOriginInsecureSkipVerify = "origin.insecureSkipVerify"

	configReplicaURL                = "replica.url"
	configReplicaWebURL             = "replica.webURL"
	configReplicaAPIPath            = "replica.apiPath"
	configReplicaUsername           = "replica.username"
	configReplicaPassword           = "replica.password"
	configReplicaCookie             = "replica.cookie"
	configReplicaInsecureSkipVerify = "replica.insecureSkipVerify"
	configReplicaAutoSetup          = "replica.autoSetup"
	configReplicaInterfaceName      = "replica.interfaceName"

	envReplicasUsernameFormat           = "REPLICA%s_USERNAME" // #nosec G101
	envReplicasPasswordFormat           = "REPLICA%s_PASSWORD" // #nosec G101
	envReplicasCookieFormat             = "REPLICA%s_COOKIE"   // #nosec G101
	envReplicasAPIPathFormat            = "REPLICA%s_APIPATH"
	envReplicasInsecureSkipVerifyFormat = "REPLICA%s_INSECURESKIPVERIFY"
	envReplicasAutoSetup                = "REPLICA%s_AUTOSETUP"
	envReplicasInterfaceName            = "REPLICA%s_INTERFACENAME"
	envDHCPServerEnabled                = "REPLICA%s_DHCPSERVERENABLED"
	envWebURL                           = "REPLICA%s_WEBURL"
)

var (
	cfgFile               string
	logger                = log.GetLogger("root")
	envReplicasURLPattern = regexp.MustCompile(`^REPLICA(\d+)_URL=(.*)`)
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "adguardhome-sync",
	Short:   "Synchronize config from one AdGuardHome instance to another",
	Version: version.Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.adguardhome-sync.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".adguardhome-sync" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".adguardhome-sync")
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger.Info("Using config file:", viper.ConfigFileUsed())
	} else if cfgFile != "" {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getConfig() (*types.Config, error) {
	cfg := &types.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}
	if cfg.Replica != nil &&
		cfg.Replica.URL == "" &&
		cfg.Replica.Username == "" {
		cfg.Replica = nil
	}

	if len(cfg.Replicas) == 0 {
		cfg.Replicas = append(cfg.Replicas, collectEnvReplicas()...)
	}

	handleDeprecatedEnvVars(cfg)

	return cfg, nil
}

func handleDeprecatedEnvVars(cfg *types.Config) {
	runOnStart := checkDeprecatedEnvVar("RUNONSTART", "RUN_ON_START")
	if runOnStart != "" {
		cfg.RunOnStart, _ = strconv.ParseBool(runOnStart)
	}
	darkMode := checkDeprecatedEnvVar("API_DARKMODE", "API_DARK_MODE")
	if darkMode != "" {
		cfg.API.DarkMode, _ = strconv.ParseBool(darkMode)
	}
	general := checkDeprecatedEnvVar("FEATURES_GENERALSETTINGS", "FEATURES_GENERAL_SETTINGS")
	if general != "" {
		cfg.Features.GeneralSettings, _ = strconv.ParseBool(general)
	}
	qlc := checkDeprecatedEnvVar("FEATURES_QUERYLOGCONFIG", "FEATURES_QUERY_LOG_CONFIG")
	if qlc != "" {
		cfg.Features.QueryLogConfig, _ = strconv.ParseBool(qlc)
	}
	stats := checkDeprecatedEnvVar("FEATURES_STATSCONFIG", "FEATURES_STATS_CONFIG")
	if stats != "" {
		cfg.Features.StatsConfig, _ = strconv.ParseBool(stats)
	}
	client := checkDeprecatedEnvVar("FEATURES_CLIENTSETTINGS", "FEATURES_CLIENT_SETTINGS")
	if client != "" {
		cfg.Features.ClientSettings, _ = strconv.ParseBool(client)
	}
	dhcpServerConfig := checkDeprecatedEnvVar("FEATURES_DHCP_SERVERCONFIG", "FEATURES_DHCP_SERVER_CONFIG")
	if dhcpServerConfig != "" {
		cfg.Features.DHCP.ServerConfig, _ = strconv.ParseBool(dhcpServerConfig)
	}
	dhcpStaticLeases := checkDeprecatedEnvVar("FEATURES_DHCP_STATICLEASES", "FEATURES_DHCP_STATIC_LEASES")
	if dhcpStaticLeases != "" {
		cfg.Features.DHCP.StaticLeases, _ = strconv.ParseBool(dhcpStaticLeases)
	}
	dnsAccessLists := checkDeprecatedEnvVar("FEATURES_DNS_ACCESSLISTS", "FEATURES_DNS_ACCESS_LISTS")
	if dnsAccessLists != "" {
		cfg.Features.DNS.AccessLists, _ = strconv.ParseBool(dnsAccessLists)
	}
	dnsServerConfig := checkDeprecatedEnvVar("FEATURES_DNS_SERVERCONFIG", "FEATURES_DNS_SERVER_CONFIG")
	if dnsServerConfig != "" {
		cfg.Features.DNS.ServerConfig, _ = strconv.ParseBool(dnsServerConfig)
	}
}

func checkDeprecatedEnvVar(oldName string, newName string) string {
	old, oldOK := os.LookupEnv(oldName)
	if !oldOK {
		return ""
	}
	logger.With("correct", newName, "deprecated", oldName).
		Warn("Deprecated env variable is used, please use the correct one")
	_, newOK := os.LookupEnv(newName)
	if newOK {
		return ""
	}
	return old
}

// Manually collect replicas from env.
func collectEnvReplicas() []types.AdGuardInstance {
	var replicas []types.AdGuardInstance
	for _, v := range os.Environ() {
		if envReplicasURLPattern.MatchString(v) {
			sm := envReplicasURLPattern.FindStringSubmatch(v)
			re := types.AdGuardInstance{
				URL:                sm[2],
				WebURL:             os.Getenv(fmt.Sprintf(envWebURL, sm[1])),
				Username:           os.Getenv(fmt.Sprintf(envReplicasUsernameFormat, sm[1])),
				Password:           os.Getenv(fmt.Sprintf(envReplicasPasswordFormat, sm[1])),
				Cookie:             os.Getenv(fmt.Sprintf(envReplicasCookieFormat, sm[1])),
				APIPath:            os.Getenv(fmt.Sprintf(envReplicasAPIPathFormat, sm[1])),
				InsecureSkipVerify: strings.EqualFold(os.Getenv(fmt.Sprintf(envReplicasInsecureSkipVerifyFormat, sm[1])), "true"),
				AutoSetup:          strings.EqualFold(os.Getenv(fmt.Sprintf(envReplicasAutoSetup, sm[1])), "true"),
				InterfaceName:      os.Getenv(fmt.Sprintf(envReplicasInterfaceName, sm[1])),
			}

			if dhcpEnabled, ok := os.LookupEnv(fmt.Sprintf(envDHCPServerEnabled, sm[1])); ok {
				if strings.EqualFold(dhcpEnabled, "true") {
					re.DHCPServerEnabled = boolPtr(true)
				} else if strings.EqualFold(dhcpEnabled, "false") {
					re.DHCPServerEnabled = boolPtr(false)
				}
			}
			if re.APIPath == "" {
				re.APIPath = "/control"
			}
			replicas = append(replicas, re)
		}
	}

	return replicas
}

func boolPtr(b bool) *bool {
	return &b
}
