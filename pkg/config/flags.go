package config

import (
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/spf13/cobra"
)

func readFlags(cfg *types.Config, cmd *cobra.Command) (err error) {
	if cmd == nil {
		return
	}

	if err = readRootFlags(cfg, cmd); err != nil {
		return err
	}

	if err = readApiFlags(cfg, cmd); err != nil {
		return err
	}

	if err = readFeatureFlags(cfg, cmd); err != nil {
		return err
	}

	if err = readOriginFlags(cfg, cmd); err != nil {
		return err
	}

	if err = readReplicaFlags(cfg, cmd); err != nil {
		return err
	}

	return
}

func readReplicaFlags(cfg *types.Config, cmd *cobra.Command) error {
	if err := setStringFlag(cfg, cmd, FlagReplicaURL, func(cgf *types.Config, value string) {
		cfg.Replica.URL = value
	}); err != nil {
		return err
	}
	if err := setStringFlag(cfg, cmd, FlagOriginWebURL, func(cgf *types.Config, value string) {
		cfg.Replica.WebURL = value
	}); err != nil {
		return err
	}
	if err := setStringFlag(cfg, cmd, FlagReplicaApiPath, func(cgf *types.Config, value string) {
		cfg.Replica.APIPath = value
	}); err != nil {
		return err
	}
	if err := setStringFlag(cfg, cmd, FlagReplicaUsername, func(cgf *types.Config, value string) {
		cfg.Replica.Username = value
	}); err != nil {
		return err
	}
	if err := setStringFlag(cfg, cmd, FlagReplicaPassword, func(cgf *types.Config, value string) {
		cfg.Replica.Password = value
	}); err != nil {
		return err
	}
	if err := setStringFlag(cfg, cmd, FlagReplicaCookie, func(cgf *types.Config, value string) {
		cfg.Replica.Cookie = value
	}); err != nil {
		return err
	}
	if err := setBoolFlag(cfg, cmd, FlagReplicaISV, func(cgf *types.Config, value bool) {
		cfg.Replica.InsecureSkipVerify = value
	}); err != nil {
		return err
	}
	if err := setBoolFlag(cfg, cmd, FlagReplicaAutoSetup, func(cgf *types.Config, value bool) {
		cfg.Replica.AutoSetup = value
	}); err != nil {
		return err
	}
	if err := setStringFlag(cfg, cmd, FlagReplicaInterfaceName, func(cgf *types.Config, value string) {
		cfg.Replica.InterfaceName = value
	}); err != nil {
		return err
	}
	return nil
}

func readOriginFlags(cfg *types.Config, cmd *cobra.Command) error {
	if err := setStringFlag(cfg, cmd, FlagOriginURL, func(cgf *types.Config, value string) {
		cfg.Origin.URL = value
	}); err != nil {
		return err
	}
	if err := setStringFlag(cfg, cmd, FlagOriginWebURL, func(cgf *types.Config, value string) {
		cfg.Origin.WebURL = value
	}); err != nil {
		return err
	}
	if err := setStringFlag(cfg, cmd, FlagOriginApiPath, func(cgf *types.Config, value string) {
		cfg.Origin.APIPath = value
	}); err != nil {
		return err
	}
	if err := setStringFlag(cfg, cmd, FlagOriginUsername, func(cgf *types.Config, value string) {
		cfg.Origin.Username = value
	}); err != nil {
		return err
	}
	if err := setStringFlag(cfg, cmd, FlagOriginPassword, func(cgf *types.Config, value string) {
		cfg.Origin.Password = value
	}); err != nil {
		return err
	}
	if err := setStringFlag(cfg, cmd, FlagOriginCookie, func(cgf *types.Config, value string) {
		cfg.Origin.Cookie = value
	}); err != nil {
		return err
	}
	if err := setBoolFlag(cfg, cmd, FlagOriginISV, func(cgf *types.Config, value bool) {
		cfg.Origin.InsecureSkipVerify = value
	}); err != nil {
		return err
	}
	return nil
}

func readFeatureFlags(cfg *types.Config, cmd *cobra.Command) error {
	if err := setBoolFlag(cfg, cmd, FlagFeatureDhcpServerConfig, func(cgf *types.Config, value bool) {
		cfg.Features.DHCP.ServerConfig = value
	}); err != nil {
		return err
	}
	if err := setBoolFlag(cfg, cmd, FlagFeatureDhcpStaticLeases, func(cgf *types.Config, value bool) {
		cfg.Features.DHCP.StaticLeases = value
	}); err != nil {
		return err
	}

	if err := setBoolFlag(cfg, cmd, FlagFeatureDnsServerConfig, func(cgf *types.Config, value bool) {
		cfg.Features.DNS.ServerConfig = value
	}); err != nil {
		return err
	}
	if err := setBoolFlag(cfg, cmd, FlagFeatureDnsAccessLists, func(cgf *types.Config, value bool) {
		cfg.Features.DNS.AccessLists = value
	}); err != nil {
		return err
	}
	if err := setBoolFlag(cfg, cmd, FlagFeatureDnsRewrites, func(cgf *types.Config, value bool) {
		cfg.Features.DNS.Rewrites = value
	}); err != nil {
		return err
	}

	if err := setBoolFlag(cfg, cmd, FlagFeatureGeneral, func(cgf *types.Config, value bool) {
		cfg.Features.GeneralSettings = value
	}); err != nil {
		return err
	}
	if err := setBoolFlag(cfg, cmd, FlagFeatureQueryLog, func(cgf *types.Config, value bool) {
		cfg.Features.QueryLogConfig = value
	}); err != nil {
		return err
	}
	if err := setBoolFlag(cfg, cmd, FlagFeatureStats, func(cgf *types.Config, value bool) {
		cfg.Features.StatsConfig = value
	}); err != nil {
		return err
	}
	if err := setBoolFlag(cfg, cmd, FlagFeatureClient, func(cgf *types.Config, value bool) {
		cfg.Features.ClientSettings = value
	}); err != nil {
		return err
	}
	if err := setBoolFlag(cfg, cmd, FlagFeatureServices, func(cgf *types.Config, value bool) {
		cfg.Features.Services = value
	}); err != nil {
		return err
	}
	if err := setBoolFlag(cfg, cmd, FlagFeatureFilters, func(cgf *types.Config, value bool) {
		cfg.Features.Filters = value
	}); err != nil {
		return err
	}

	return nil
}

func readApiFlags(cfg *types.Config, cmd *cobra.Command) (err error) {
	if err = setIntFlag(cfg, cmd, FlagApiPort, func(cgf *types.Config, value int) {
		cfg.API.Port = value
	}); err != nil {
		return
	}
	if err = setStringFlag(cfg, cmd, FlagApiUsername, func(cgf *types.Config, value string) {
		cfg.API.Username = value
	}); err != nil {
		return
	}
	if err = setStringFlag(cfg, cmd, FlagApiPassword, func(cgf *types.Config, value string) {
		cfg.API.Password = value
	}); err != nil {
		return
	}
	if err = setBoolFlag(cfg, cmd, FlagApiDarkMode, func(cgf *types.Config, value bool) {
		cfg.API.DarkMode = value
	}); err != nil {
		return
	}
	return
}

func readRootFlags(cfg *types.Config, cmd *cobra.Command) (err error) {
	if err = setStringFlag(cfg, cmd, FlagCron, func(cgf *types.Config, value string) {
		cfg.Cron = value
	}); err != nil {
		return
	}
	if err = setBoolFlag(cfg, cmd, FlagRunOnStart, func(cgf *types.Config, value bool) {
		cfg.RunOnStart = value
	}); err != nil {
		return
	}
	if err = setBoolFlag(cfg, cmd, FlagPrintConfigOnly, func(cgf *types.Config, value bool) {
		cfg.PrintConfigOnly = value
	}); err != nil {
		return
	}
	if err = setBoolFlag(cfg, cmd, FlagContinueOnError, func(cgf *types.Config, value bool) {
		cfg.ContinueOnError = value
	}); err != nil {
		return
	}
	return
}

func setStringFlag(cfg *types.Config, cmd *cobra.Command, name string, cb callback[string]) (err error) {
	if cmd.Flags().Changed(name) {
		if value, err := cmd.Flags().GetString(name); err != nil {
			return err
		} else {
			cb(cfg, value)
		}
	}
	return nil
}

func setBoolFlag(cfg *types.Config, cmd *cobra.Command, name string, cb callback[bool]) (err error) {
	if cmd.Flags().Changed(name) {
		if value, err := cmd.Flags().GetBool(name); err != nil {
			return err
		} else {
			cb(cfg, value)
		}
	}
	return nil
}

func setIntFlag(cfg *types.Config, cmd *cobra.Command, name string, cb callback[int]) (err error) {
	if cmd.Flags().Changed(name) {
		if value, err := cmd.Flags().GetInt(name); err != nil {
			return err
		} else {
			cb(cfg, value)
		}
	}
	return nil
}

type callback[T any] func(cgf *types.Config, value T)
