package cmd

import (
	"github.com/bakito/adguardhome-sync/pkg/config"
	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/sync"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// runCmd represents the run command
var doCmd = &cobra.Command{
	Use:   "run",
	Short: "Start a synchronisation from origin to replica",
	Long:  `Synchronizes the configuration form an origin instance to a replica`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger = log.GetLogger("run")
		cfg, err := config.Get(cfgFile, cmd.Flags())
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
	doCmd.PersistentFlags().String(config.FlagCron, "", "The cron expression to run in daemon mode")
	doCmd.PersistentFlags().Bool(config.FlagRunOnStart, true, "Run the sync job on start.")
	doCmd.PersistentFlags().Bool(config.FlagPrintConfigOnly, false, "Prints the configuration only and exists. "+
		"Can be used to debug the config E.g: when having authentication issues.")
	doCmd.PersistentFlags().Bool(config.FlagContinueOnError, false, "If enabled, the synchronisation task "+
		"will not fail on single errors, but will log the errors and continue.")

	doCmd.PersistentFlags().
		Int(config.FlagApiPort, 8080, "Sync API Port, the API endpoint will be started to enable remote triggering; if 0 port API is disabled.")
	doCmd.PersistentFlags().String(config.FlagApiUsername, "", "Sync API username")
	doCmd.PersistentFlags().String(config.FlagApiPassword, "", "Sync API password")
	doCmd.PersistentFlags().String(config.FlagApiDarkMode, "", "API UI in dark mode")

	doCmd.PersistentFlags().Bool(config.FlagFeatureDhcpServerConfig, true, "Enable DHCP server config feature")
	doCmd.PersistentFlags().Bool(config.FlagFeatureDhcpStaticLeases, true, "Enable DHCP server static leases feature")

	doCmd.PersistentFlags().Bool(config.FlagFeatureDnsServerConfig, true, "Enable DNS server config feature")
	doCmd.PersistentFlags().Bool(config.FlagFeatureDnsAccessLists, true, "Enable DNS server access lists feature")
	doCmd.PersistentFlags().Bool(config.FlagFeatureDnsRewrites, true, "Enable DNS rewrites feature")

	doCmd.PersistentFlags().Bool(config.FlagFeatureGeneral, true, "Enable general settings feature")
	doCmd.PersistentFlags().Bool(config.FlagFeatureQueryLog, true, "Enable query log config feature")
	doCmd.PersistentFlags().Bool(config.FlagFeatureStats, true, "Enable stats config feature")
	doCmd.PersistentFlags().Bool(config.FlagFeatureClient, true, "Enable client settings feature")
	doCmd.PersistentFlags().Bool(config.FlagFeatureServices, true, "Enable services sync feature")
	doCmd.PersistentFlags().Bool(config.FlagFeatureFilters, true, "Enable filters sync feature")

	doCmd.PersistentFlags().String(config.FlagOriginURL, "", "Origin instance url")
	doCmd.PersistentFlags().
		String(config.FlagOriginWebURL, "", "Origin instance web url used in the web interface (default: <origin-url>)")
	doCmd.PersistentFlags().String(config.FlagOriginApiPath, "/control", "Origin instance API path")
	doCmd.PersistentFlags().String(config.FlagOriginUsername, "", "Origin instance username")
	doCmd.PersistentFlags().String(config.FlagOriginPassword, "", "Origin instance password")
	doCmd.PersistentFlags().String(config.FlagOriginCookie, "", "If Set, uses a cookie for authentication")
	doCmd.PersistentFlags().Bool(config.FlagOriginISV, false, "Enable Origin instance InsecureSkipVerify")

	doCmd.PersistentFlags().String(config.FlagReplicaURL, "", "Replica instance url")
	doCmd.PersistentFlags().
		String(config.FlagReplicaWebURL, "", "Replica instance web url used in the web interface (default: <replica-url>)")
	doCmd.PersistentFlags().String(config.FlagReplicaApiPath, "/control", "Replica instance API path")
	doCmd.PersistentFlags().String(config.FlagReplicaUsername, "", "Replica instance username")
	doCmd.PersistentFlags().String(config.FlagReplicaPassword, "", "Replica instance password")
	doCmd.PersistentFlags().String(config.FlagReplicaCookie, "", "If Set, uses a cookie for authentication")
	doCmd.PersistentFlags().Bool(config.FlagReplicaISV, false, "Enable Replica instance InsecureSkipVerify")
	doCmd.PersistentFlags().
		Bool(config.FlagReplicaAutoSetup, false, "Enable automatic setup of new AdguardHome instances. This replaces the setup wizard.")
	doCmd.PersistentFlags().
		String(config.FlagReplicaInterfaceName, "", "Optional change the interface name of the replica if it differs from the master")
}
