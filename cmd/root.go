package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/bakito/adguardhome-sync/version"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	configCron       = "cron"
	configRunOnStart = "runOnStart"

	configAPIPort     = "api.port"
	configAPIUsername = "api.username"
	configAPIPassword = "api.password"
	configAPIDarkMode = "api.darkMode"

	configFeatureDHCPServerConfig = "features.dhcp.serverConfig"
	configFeatureDHCPStaticLeases = "features.dhcp.staticLeases"
	configFeatureDNServerConfig   = "features.dns.serverConfig"
	configFeatureDNSPAccessLists  = "features.dns.accessLists"
	configFeatureDNSRewrites      = "features.dns.rewrites"
	configFeatureGeneralSettings  = "features.generalSettings"
	configFeatureQueryLogConfig   = "features.queryLogConfig"
	configFeatureStatsConfig      = "features.statsConfig"
	configFeatureClientSettings   = "features.clientSettings"
	configFeatureServices         = "features.services"
	configFeatureFilters          = "features.filters"

	configOriginURL                = "origin.url"
	configOriginAPIPath            = "origin.apiPath"
	configOriginUsername           = "origin.username"
	configOriginPassword           = "origin.password"
	configOriginInsecureSkipVerify = "origin.insecureSkipVerify"

	configReplicaURL                = "replica.url"
	configReplicaAPIPath            = "replica.apiPath"
	configReplicaUsername           = "replica.username"
	configReplicaPassword           = "replica.password"
	configReplicaInsecureSkipVerify = "replica.insecureSkipVerify"
	configReplicaAutoSetup          = "replica.autoSetup"
	configReplicaInterfaceName      = "replica.interfaceName"

	envReplicasUsernameFormat           = "REPLICA%s_USERNAME" // #nosec G101
	envReplicasPasswordFormat           = "REPLICA%s_PASSWORD" // #nosec G101
	envReplicasAPIPathFormat            = "REPLICA%s_APIPATH"
	envReplicasInsecureSkipVerifyFormat = "REPLICA%s_INSECURESKIPVERIFY"
	envReplicasAutoSetup                = "REPLICA%s_AUTOSETUP"
	envReplicasInterfaceName            = "REPLICA%s_INTERFACENAME"
	// Deprecated: use envReplicasInterfaceName instead
	envReplicasInterfaceNameDeprecated = "REPLICA%s_INTERFACWENAME"
	envDHCPServerEnabled               = "REPLICA%s_DHCPSERVERENABLED"
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

func getConfig(logger *zap.SugaredLogger) (*types.Config, error) {
	cfg := &types.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	if len(cfg.Replicas) == 0 {
		cfg.Replicas = append(cfg.Replicas, collectEnvReplicas(logger)...)
	}
	return cfg, nil
}

// Manually collect replicas from env.
func collectEnvReplicas(logger *zap.SugaredLogger) []types.AdGuardInstance {
	var replicas []types.AdGuardInstance
	for _, v := range os.Environ() {
		if envReplicasURLPattern.MatchString(v) {
			sm := envReplicasURLPattern.FindStringSubmatch(v)
			re := types.AdGuardInstance{
				URL:                sm[2],
				Username:           os.Getenv(fmt.Sprintf(envReplicasUsernameFormat, sm[1])),
				Password:           os.Getenv(fmt.Sprintf(envReplicasPasswordFormat, sm[1])),
				APIPath:            os.Getenv(fmt.Sprintf(envReplicasAPIPathFormat, sm[1])),
				InsecureSkipVerify: strings.EqualFold(os.Getenv(fmt.Sprintf(envReplicasInsecureSkipVerifyFormat, sm[1])), "true"),
				AutoSetup:          strings.EqualFold(os.Getenv(fmt.Sprintf(envReplicasAutoSetup, sm[1])), "true"),
				InterfaceName:      os.Getenv(fmt.Sprintf(envReplicasInterfaceName, sm[1])),
			}
			if re.InterfaceName != "" {
				if in, ok := os.LookupEnv(fmt.Sprintf(envReplicasInterfaceNameDeprecated, sm[1])); ok {
					logger.
						With("correct", envReplicasInterfaceName, "deprecated", envReplicasInterfaceNameDeprecated).
						Warn("Deprecated env variable is used, please use the correct one")
					re.InterfaceName = in
				}
			}
			if dhcpEnabled, ok := os.LookupEnv(envDHCPServerEnabled); ok {
				if strings.EqualFold(dhcpEnabled, "true") {
					re.DHCPServerEnabled = boolPtr(true)
				} else if strings.EqualFold(dhcpEnabled, "false") {
					re.DHCPServerEnabled = boolPtr(false)
				}
			}
			replicas = append(replicas, re)
		}
	}

	return replicas
}

func boolPtr(b bool) *bool {
	return &b
}
