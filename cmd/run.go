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
		cfg, err := config.Get(cfgFile)
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
	doCmd.PersistentFlags().Bool("runOnStart", true, "Run the sync job on start.")
	doCmd.PersistentFlags().Bool("printConfigOnly", false, "Prints the configuration only and exists. "+
		"Can be used to debug the config E.g: when having authentication issues.")
	doCmd.PersistentFlags().Bool("continueOnError", false, "If enabled, the synchronisation task "+
		"will not fail on single errors, but will log the errors and continue.")

	doCmd.PersistentFlags().Int("api-port", 8080, "Sync API Port, the API endpoint will be started to enable remote triggering; if 0 port API is disabled.")
	doCmd.PersistentFlags().String("api-username", "", "Sync API username")
	doCmd.PersistentFlags().String("api-password", "", "Sync API password")
	doCmd.PersistentFlags().String("api-darkMode", "", "API UI in dark mode")

	doCmd.PersistentFlags().Bool("feature-dhcp-server-config", true, "Enable DHCP server config feature")
	doCmd.PersistentFlags().Bool("feature-dhcp-static-leases", true, "Enable DHCP server static leases feature")

	doCmd.PersistentFlags().Bool("feature-dns-server-config", true, "Enable DNS server config feature")
	doCmd.PersistentFlags().Bool("feature-dns-access-lists", true, "Enable DNS server access lists feature")
	doCmd.PersistentFlags().Bool("feature-dns-rewrites", true, "Enable DNS rewrites feature")
	doCmd.PersistentFlags().Bool("feature-general-settings", true, "Enable general settings feature")
	doCmd.PersistentFlags().Bool("feature-query-log-config", true, "Enable query log config feature")
	doCmd.PersistentFlags().Bool("feature-stats-config", true, "Enable stats config feature")
	doCmd.PersistentFlags().Bool("feature-client-settings", true, "Enable client settings feature")
	doCmd.PersistentFlags().Bool("feature-services", true, "Enable services sync feature")
	doCmd.PersistentFlags().Bool("feature-filters", true, "Enable filters sync feature")

	doCmd.PersistentFlags().String("origin-url", "", "Origin instance url")
	doCmd.PersistentFlags().String("origin-web-url", "", "Origin instance web url used in the web interface (default: <origin-url>)")
	doCmd.PersistentFlags().String("origin-api-path", "/control", "Origin instance API path")
	doCmd.PersistentFlags().String("origin-username", "", "Origin instance username")
	doCmd.PersistentFlags().String("origin-password", "", "Origin instance password")
	doCmd.PersistentFlags().String("origin-cookie", "", "If Set, uses a cookie for authentication")
	doCmd.PersistentFlags().Bool("origin-insecure-skip-verify", false, "Enable Origin instance InsecureSkipVerify")

	doCmd.PersistentFlags().String("replica-url", "", "Replica instance url")
	doCmd.PersistentFlags().String("replica-web-url", "", "Replica instance web url used in the web interface (default: <replica-url>)")
	doCmd.PersistentFlags().String("replica-api-path", "/control", "Replica instance API path")
	doCmd.PersistentFlags().String("replica-username", "", "Replica instance username")
	doCmd.PersistentFlags().String("replica-password", "", "Replica instance password")
	doCmd.PersistentFlags().String("replica-cookie", "", "If Set, uses a cookie for authentication")
	doCmd.PersistentFlags().Bool("replica-insecure-skip-verify", false, "Enable Replica instance InsecureSkipVerify")
	doCmd.PersistentFlags().Bool("replica-auto-setup", false, "Enable automatic setup of new AdguardHome instances. This replaces the setup wizard.")
	doCmd.PersistentFlags().String("replica-interface-name", "", "Optional change the interface name of the replica if it differs from the master")
}
