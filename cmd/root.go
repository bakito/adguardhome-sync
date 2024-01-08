package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/bakito/adguardhome-sync/pkg/utils"
	"github.com/bakito/adguardhome-sync/version"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	configCron            = "CRON"
	configRunOnStart      = "RUN_ON_START"
	configPrintConfigOnly = "PRINT_CONFIG_ONLY"
	configContinueOnError = "CONTINUE_ON_ERROR"

	configAPIPort     = "API.PORT"
	configAPIUsername = "API.USERNAME"
	configAPIPassword = "API.PASSWORD"
	configAPIDarkMode = "API.DARK_MODE"

	configFeatureDHCPServerConfig = "FEATURES.DHCP.SERVER_CONFIG"
	configFeatureDHCPStaticLeases = "FEATURES.DHCP.STATIC_LEASES"
	configFeatureDNServerConfig   = "FEATURES.DNS.SERVER_CONFIG"
	configFeatureDNSPAccessLists  = "FEATURES.DNS.ACCESS_LISTS"
	configFeatureDNSRewrites      = "FEATURES.DNS.rewrites"
	configFeatureGeneralSettings  = "FEATURES.GENERAL_SETTINGS"
	configFeatureQueryLogConfig   = "FEATURES.QUERY_LOG_CONFIG"
	configFeatureStatsConfig      = "FEATURES.STATS_CONFIG"
	configFeatureClientSettings   = "FEATURES.CLIENT_SETTINGS"
	configFeatureServices         = "FEATURES.SERVICES"
	configFeatureFilters          = "FEATURES.FILTERS"

	configOriginURL                = "ORIGIN.URL"
	configOriginWebURL             = "ORIGIN.WEB_URL"
	configOriginAPIPath            = "ORIGIN.API_PATH"
	configOriginUsername           = "ORIGIN.USERNAME"
	configOriginPassword           = "ORIGIN.PASSWORD"
	configOriginCookie             = "ORIGIN.COOKIE"
	configOriginInsecureSkipVerify = "ORIGIN.INSECURE_SKIP_VERIFY"

	configReplicaURL                = "REPLICA.URL"
	configReplicaWebURL             = "REPLICA.WEB_URL"
	configReplicaAPIPath            = "REPLICA.API_PATH"
	configReplicaUsername           = "REPLICA.USERNAME"
	configReplicaPassword           = "REPLICA.PASSWORD"
	configReplicaCookie             = "REPLICA.COOKIE"
	configReplicaInsecureSkipVerify = "REPLICA.INSECURE_SKIP_VERIFY"
	configReplicaAutoSetup          = "REPLICA.AUTO_SETUP"
	configReplicaInterfaceName      = "REPLICA.INTERFACE_NAME"
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
	value := checkDeprecatedEnvVar("RUNONSTART", "RUN_ON_START")
	if value != "" {
		cfg.RunOnStart, _ = strconv.ParseBool(value)
	}
	value = checkDeprecatedEnvVar("API_DARKMODE", "API_DARK_MODE")
	if value != "" {
		cfg.API.DarkMode, _ = strconv.ParseBool(value)
	}
	value = checkDeprecatedEnvVar("FEATURES_GENERALSETTINGS", "FEATURES_GENERAL_SETTINGS")
	if value != "" {
		cfg.Features.GeneralSettings, _ = strconv.ParseBool(value)
	}
	value = checkDeprecatedEnvVar("FEATURES_QUERYLOGCONFIG", "FEATURES_QUERY_LOG_CONFIG")
	if value != "" {
		cfg.Features.QueryLogConfig, _ = strconv.ParseBool(value)
	}
	value = checkDeprecatedEnvVar("FEATURES_STATSCONFIG", "FEATURES_STATS_CONFIG")
	if value != "" {
		cfg.Features.StatsConfig, _ = strconv.ParseBool(value)
	}
	value = checkDeprecatedEnvVar("FEATURES_CLIENTSETTINGS", "FEATURES_CLIENT_SETTINGS")
	if value != "" {
		cfg.Features.ClientSettings, _ = strconv.ParseBool(value)
	}
	value = checkDeprecatedEnvVar("FEATURES_DHCP_SERVERCONFIG", "FEATURES_DHCP_SERVER_CONFIG")
	if value != "" {
		cfg.Features.DHCP.ServerConfig, _ = strconv.ParseBool(value)
	}
	value = checkDeprecatedEnvVar("FEATURES_DHCP_STATICLEASES", "FEATURES_DHCP_STATIC_LEASES")
	if value != "" {
		cfg.Features.DHCP.StaticLeases, _ = strconv.ParseBool(value)
	}
	value = checkDeprecatedEnvVar("FEATURES_DNS_ACCESSLISTS", "FEATURES_DNS_ACCESS_LISTS")
	if value != "" {
		cfg.Features.DNS.AccessLists, _ = strconv.ParseBool(value)
	}
	value = checkDeprecatedEnvVar("FEATURES_DNS_SERVERCONFIG", "FEATURES_DNS_SERVER_CONFIG")
	if value != "" {
		cfg.Features.DNS.ServerConfig, _ = strconv.ParseBool(value)
	}

	if cfg.Replica != nil {
		value = checkDeprecatedEnvVar("REPLICA_WEBURL", "REPLICA_WEB_URL")
		if value != "" {
			cfg.Replica.WebURL = value
		}
		value = checkDeprecatedEnvVar("REPLICA_AUTOSETUP", "REPLICA_AUTO_SETUP")
		if value != "" {
			cfg.Replica.AutoSetup, _ = strconv.ParseBool(value)
		}
		value = checkDeprecatedEnvVar("REPLICA_INTERFACENAME", "REPLICA_INTERFACE_NAME")
		if value != "" {
			cfg.Replica.InterfaceName = value
		}
		value = checkDeprecatedEnvVar("REPLICA_DHCPSERVERENABLED", "REPLICA_DHCP_SERVER_ENABLED")
		if value != "" {
			if b, err := strconv.ParseBool(value); err != nil {
				cfg.Replica.DHCPServerEnabled = utils.Ptr(b)
			}
		}
	}
}

func checkDeprecatedEnvVar(oldName string, newName string) string {
	old, oldOK := os.LookupEnv(oldName)
	if oldOK {
		logger.With("deprecated", oldName, "replacement", newName).
			Warn("Deprecated env variable is used, please use the correct one")
	}
	new, newOK := os.LookupEnv(newName)
	if newOK {
		return new
	}
	return old
}

func checkDeprecatedReplicaEnvVar(oldPattern, newPattern, replicaID string) string {
	return checkDeprecatedEnvVar(fmt.Sprintf(oldPattern, replicaID), fmt.Sprintf(newPattern, replicaID))
}

// Manually collect replicas from env.
func collectEnvReplicas() []types.AdGuardInstance {
	var replicas []types.AdGuardInstance
	for _, v := range os.Environ() {
		if envReplicasURLPattern.MatchString(v) {
			sm := envReplicasURLPattern.FindStringSubmatch(v)
			index := sm[1]
			re := types.AdGuardInstance{
				URL:                sm[2],
				WebURL:             os.Getenv(fmt.Sprintf("REPLICA%s_WEB_URL", index)),
				APIPath:            checkDeprecatedReplicaEnvVar("REPLICA%s_APIPATH", "REPLICA%s_API_PATH", index),
				Username:           os.Getenv(fmt.Sprintf("REPLICA%s_USERNAME", index)),
				Password:           os.Getenv(fmt.Sprintf("REPLICA%s_PASSWORD", index)),
				Cookie:             os.Getenv(fmt.Sprintf("REPLICA%s_COOKIE", index)),
				InsecureSkipVerify: strings.EqualFold(checkDeprecatedReplicaEnvVar("REPLICA%s_INSECURESKIPVERIFY", "REPLICA%s_INSECURE_SKIP_VERIFY", index), "true"),
				AutoSetup:          strings.EqualFold(checkDeprecatedReplicaEnvVar("REPLICA%s_AUTOSETUP", "REPLICA%s_AUTO_SETUP", index), "true"),
				InterfaceName:      checkDeprecatedReplicaEnvVar("REPLICA%s_INTERFACENAME", "REPLICA%s_INTERFACE_NAME", index),
			}

			dhcpEnabled := checkDeprecatedReplicaEnvVar("REPLICA%s_DHCPSERVERENABLED", "REPLICA%s_DHCP_SERVER_ENABLED", index)
			if strings.EqualFold(dhcpEnabled, "true") {
				re.DHCPServerEnabled = utils.Ptr(true)
			} else if strings.EqualFold(dhcpEnabled, "false") {
				re.DHCPServerEnabled = utils.Ptr(false)
			}
			if re.APIPath == "" {
				re.APIPath = "/control"
			}
			replicas = append(replicas, re)
		}
	}

	return replicas
}
