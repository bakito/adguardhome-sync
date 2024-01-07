package cmd

import (
	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/sync"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// runCmd represents the run command
var doCmd = &cobra.Command{
	Use:   "run",
	Short: "Start a synchronisation from origin to replica",
	Long:  `Synchronizes the configuration form an origin instance to a replica`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger = log.GetLogger("run")
		cfg, err := getConfig(logger)
		if err != nil {
			logger.Error(err)
			return err
		}

		if err := cfg.Init(); err != nil {
			logger.Error(err)
			return err
		}

		if cfg.PrintConfigOnly {
			config, err := yaml.Marshal(cfg)
			if err != nil {
				logger.Error(err)
				return err
			}
			logger.Infof("Printing adguardhome-sync config (THE APPLICATION WILL NOT START IN THIS MODE): \n%s",
				string(config))
			return nil
		}

		return sync.Sync(cfg)
	},
}

func init() {
	rootCmd.AddCommand(doCmd)
	doCmd.PersistentFlags().String("cron", "", "The cron expression to run in daemon mode")
	_ = viper.BindPFlag(configCron, doCmd.PersistentFlags().Lookup("cron"))
	doCmd.PersistentFlags().Bool("runOnStart", true, "Run the sync job on start.")
	_ = viper.BindPFlag(configRunOnStart, doCmd.PersistentFlags().Lookup("runOnStart"))
	doCmd.PersistentFlags().Bool("printConfigOnly", false, "Prints the configuration only and exists. "+
		"Can be used to debug the config E.g: when having authentication issues.")
	_ = viper.BindPFlag(configPrintConfigOnly, doCmd.PersistentFlags().Lookup("printConfigOnly"))
	doCmd.PersistentFlags().Int("api-port", 8080, "Sync API Port, the API endpoint will be started to enable remote triggering; if 0 port API is disabled.")
	_ = viper.BindPFlag(configAPIPort, doCmd.PersistentFlags().Lookup("api-port"))
	doCmd.PersistentFlags().String("api-username", "", "Sync API username")
	_ = viper.BindPFlag(configAPIUsername, doCmd.PersistentFlags().Lookup("api-username"))
	doCmd.PersistentFlags().String("api-password", "", "Sync API password")
	_ = viper.BindPFlag(configAPIPassword, doCmd.PersistentFlags().Lookup("api-password"))
	doCmd.PersistentFlags().String("api-darkMode", "", "API UI in dark mode")
	_ = viper.BindPFlag(configAPIDarkMode, doCmd.PersistentFlags().Lookup("api-darkMode"))

	doCmd.PersistentFlags().Bool("feature-dhcp-server-config", true, "Enable DHCP server config feature")
	_ = viper.BindPFlag(configFeatureDHCPServerConfig, doCmd.PersistentFlags().Lookup("feature-dhcp-server-config"))
	doCmd.PersistentFlags().Bool("feature-dhcp-static-leases", true, "Enable DHCP server static leases feature")
	_ = viper.BindPFlag(configFeatureDHCPStaticLeases, doCmd.PersistentFlags().Lookup("feature-dhcp-static-leases"))

	doCmd.PersistentFlags().Bool("feature-dns-server-config", true, "Enable DNS server config feature")
	_ = viper.BindPFlag(configFeatureDNServerConfig, doCmd.PersistentFlags().Lookup("feature-dns-server-config"))
	doCmd.PersistentFlags().Bool("feature-dns-access-lists", true, "Enable DNS server access lists feature")
	_ = viper.BindPFlag(configFeatureDNSPAccessLists, doCmd.PersistentFlags().Lookup("feature-dns-access-lists"))
	doCmd.PersistentFlags().Bool("feature-dns-rewrites", true, "Enable DNS rewrites feature")
	_ = viper.BindPFlag(configFeatureDNSRewrites, doCmd.PersistentFlags().Lookup("feature-dns-rewrites"))
	doCmd.PersistentFlags().Bool("feature-general-settings", true, "Enable general settings feature")
	_ = viper.BindPFlag(configFeatureGeneralSettings, doCmd.PersistentFlags().Lookup("feature-general-settings"))
	_ = viper.BindPFlag("features.generalSettings", doCmd.PersistentFlags().Lookup("feature-general-settings"))
	doCmd.PersistentFlags().Bool("feature-query-log-config", true, "Enable query log config feature")
	_ = viper.BindPFlag(configFeatureQueryLogConfig, doCmd.PersistentFlags().Lookup("feature-query-log-config"))
	doCmd.PersistentFlags().Bool("feature-stats-config", true, "Enable stats config feature")
	_ = viper.BindPFlag(configFeatureStatsConfig, doCmd.PersistentFlags().Lookup("feature-stats-config"))
	doCmd.PersistentFlags().Bool("feature-client-settings", true, "Enable client settings feature")
	_ = viper.BindPFlag(configFeatureClientSettings, doCmd.PersistentFlags().Lookup("feature-client-settings"))
	doCmd.PersistentFlags().Bool("feature-services", true, "Enable services sync feature")
	_ = viper.BindPFlag(configFeatureServices, doCmd.PersistentFlags().Lookup("feature-services"))
	doCmd.PersistentFlags().Bool("feature-filters", true, "Enable filters sync feature")
	_ = viper.BindPFlag(configFeatureFilters, doCmd.PersistentFlags().Lookup("feature-filters"))

	doCmd.PersistentFlags().String("origin-url", "", "Origin instance url")
	_ = viper.BindPFlag(configOriginURL, doCmd.PersistentFlags().Lookup("origin-url"))
	doCmd.PersistentFlags().String("origin-weburl", "", "Origin instance web url used in the web interface (default: <origin-url>)")
	_ = viper.BindPFlag(configOriginWebURL, doCmd.PersistentFlags().Lookup("origin-weburl"))
	doCmd.PersistentFlags().String("origin-api-path", "/control", "Origin instance API path")
	_ = viper.BindPFlag(configOriginAPIPath, doCmd.PersistentFlags().Lookup("origin-api-path"))
	doCmd.PersistentFlags().String("origin-username", "", "Origin instance username")
	_ = viper.BindPFlag(configOriginUsername, doCmd.PersistentFlags().Lookup("origin-username"))
	doCmd.PersistentFlags().String("origin-password", "", "Origin instance password")
	_ = viper.BindPFlag(configOriginPassword, doCmd.PersistentFlags().Lookup("origin-password"))
	doCmd.PersistentFlags().String("origin-cookie", "", "If Set, uses a cookie for authentication")
	_ = viper.BindPFlag(configOriginCookie, doCmd.PersistentFlags().Lookup("origin-cookie"))
	doCmd.PersistentFlags().Bool("origin-insecure-skip-verify", false, "Enable Origin instance InsecureSkipVerify")
	_ = viper.BindPFlag(configOriginInsecureSkipVerify, doCmd.PersistentFlags().Lookup("origin-insecure-skip-verify"))

	doCmd.PersistentFlags().String("replica-url", "", "Replica instance url")
	_ = viper.BindPFlag(configReplicaURL, doCmd.PersistentFlags().Lookup("replica-url"))
	doCmd.PersistentFlags().String("replica-weburl", "", "Replica instance web url used in the web interface (default: <replica-url>)")
	_ = viper.BindPFlag(configOriginWebURL, doCmd.PersistentFlags().Lookup("replica-weburl"))
	doCmd.PersistentFlags().String("replica-api-path", "/control", "Replica instance API path")
	_ = viper.BindPFlag(configReplicaAPIPath, doCmd.PersistentFlags().Lookup("replica-api-path"))
	doCmd.PersistentFlags().String("replica-username", "", "Replica instance username")
	_ = viper.BindPFlag(configReplicaUsername, doCmd.PersistentFlags().Lookup("replica-username"))
	doCmd.PersistentFlags().String("replica-password", "", "Replica instance password")
	_ = viper.BindPFlag(configReplicaPassword, doCmd.PersistentFlags().Lookup("replica-password"))
	doCmd.PersistentFlags().String("replica-cookie", "", "If Set, uses a cookie for authentication")
	_ = viper.BindPFlag(configReplicaCookie, doCmd.PersistentFlags().Lookup("replica-cookie"))
	doCmd.PersistentFlags().Bool("replica-insecure-skip-verify", false, "Enable Replica instance InsecureSkipVerify")
	_ = viper.BindPFlag(configReplicaInsecureSkipVerify, doCmd.PersistentFlags().Lookup("replica-insecure-skip-verify"))
	doCmd.PersistentFlags().Bool("replica-auto-setup", false, "Enable automatic setup of new AdguardHome instances. This replaces the setup wizard.")
	_ = viper.BindPFlag(configReplicaAutoSetup, doCmd.PersistentFlags().Lookup("replica-auto-setup"))
	doCmd.PersistentFlags().String("replica-interface-name", "", "Optional change the interface name of the replica if it differs from the master")
	_ = viper.BindPFlag(configReplicaInterfaceName, doCmd.PersistentFlags().Lookup("replica-interface-name"))
}
