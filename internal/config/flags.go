package config

import (
	"github.com/bakito/adguardhome-sync/internal/types"
)

func readFlags(cfg *types.Config, flags Flags) error {
	if flags == nil {
		return nil
	}

	fr := &flagReader{
		cfg:   cfg,
		flags: flags,
	}

	if err := fr.readRootFlags(); err != nil {
		return err
	}

	if err := fr.readAPIFlags(); err != nil {
		return err
	}

	if err := fr.readFeatureFlags(); err != nil {
		return err
	}

	if err := fr.readOriginFlags(); err != nil {
		return err
	}

	return fr.readReplicaFlags()
}

type flagReader struct {
	cfg   *types.Config
	flags Flags
}

func (fr *flagReader) readReplicaFlags() error {
	if err := fr.setStringFlag(FlagReplicaURL, func(cgf *types.Config, value string) {
		fr.cfg.Replica.URL = value
	}); err != nil {
		return err
	}
	if err := fr.setStringFlag(FlagReplicaWebURL, func(cgf *types.Config, value string) {
		fr.cfg.Replica.WebURL = value
	}); err != nil {
		return err
	}
	if err := fr.setStringFlag(FlagReplicaAPIPath, func(cgf *types.Config, value string) {
		fr.cfg.Replica.APIPath = value
	}); err != nil {
		return err
	}
	if err := fr.setStringFlag(FlagReplicaUsername, func(cgf *types.Config, value string) {
		fr.cfg.Replica.Username = value
	}); err != nil {
		return err
	}
	if err := fr.setStringFlag(FlagReplicaPassword, func(cgf *types.Config, value string) {
		fr.cfg.Replica.Password = value
	}); err != nil {
		return err
	}
	if err := fr.setStringFlag(FlagReplicaCookie, func(cgf *types.Config, value string) {
		fr.cfg.Replica.Cookie = value
	}); err != nil {
		return err
	}
	if err := fr.setBoolFlag(FlagReplicaISV, func(cgf *types.Config, value bool) {
		fr.cfg.Replica.InsecureSkipVerify = value
	}); err != nil {
		return err
	}
	if err := fr.setBoolFlag(FlagReplicaAutoSetup, func(cgf *types.Config, value bool) {
		fr.cfg.Replica.AutoSetup = value
	}); err != nil {
		return err
	}
	return fr.setStringFlag(FlagReplicaInterfaceName, func(cgf *types.Config, value string) {
		fr.cfg.Replica.InterfaceName = value
	})
}

func (fr *flagReader) readOriginFlags() error {
	if err := fr.setStringFlag(FlagOriginURL, func(cgf *types.Config, value string) {
		fr.cfg.Origin.URL = value
	}); err != nil {
		return err
	}
	if err := fr.setStringFlag(FlagOriginWebURL, func(cgf *types.Config, value string) {
		fr.cfg.Origin.WebURL = value
	}); err != nil {
		return err
	}
	if err := fr.setStringFlag(FlagOriginAPIPath, func(cgf *types.Config, value string) {
		fr.cfg.Origin.APIPath = value
	}); err != nil {
		return err
	}
	if err := fr.setStringFlag(FlagOriginUsername, func(cgf *types.Config, value string) {
		fr.cfg.Origin.Username = value
	}); err != nil {
		return err
	}
	if err := fr.setStringFlag(FlagOriginPassword, func(cgf *types.Config, value string) {
		fr.cfg.Origin.Password = value
	}); err != nil {
		return err
	}
	if err := fr.setStringFlag(FlagOriginCookie, func(cgf *types.Config, value string) {
		fr.cfg.Origin.Cookie = value
	}); err != nil {
		return err
	}
	return fr.setBoolFlag(FlagOriginISV, func(cgf *types.Config, value bool) {
		fr.cfg.Origin.InsecureSkipVerify = value
	})
}

func (fr *flagReader) readFeatureFlags() error {
	if err := fr.setBoolFlag(FlagFeatureDhcpServerConfig, func(cgf *types.Config, value bool) {
		fr.cfg.Features.DHCP.ServerConfig = value
	}); err != nil {
		return err
	}
	if err := fr.setBoolFlag(FlagFeatureDhcpStaticLeases, func(cgf *types.Config, value bool) {
		fr.cfg.Features.DHCP.StaticLeases = value
	}); err != nil {
		return err
	}

	if err := fr.setBoolFlag(FlagFeatureDNSServerConfig, func(cgf *types.Config, value bool) {
		fr.cfg.Features.DNS.ServerConfig = value
	}); err != nil {
		return err
	}
	if err := fr.setBoolFlag(FlagFeatureDNSAccessLists, func(cgf *types.Config, value bool) {
		fr.cfg.Features.DNS.AccessLists = value
	}); err != nil {
		return err
	}
	if err := fr.setBoolFlag(FlagFeatureDNSRewrites, func(cgf *types.Config, value bool) {
		fr.cfg.Features.DNS.Rewrites = value
	}); err != nil {
		return err
	}

	if err := fr.setBoolFlag(FlagFeatureGeneral, func(cgf *types.Config, value bool) {
		fr.cfg.Features.GeneralSettings = value
	}); err != nil {
		return err
	}
	if err := fr.setBoolFlag(FlagFeatureQueryLog, func(cgf *types.Config, value bool) {
		fr.cfg.Features.QueryLogConfig = value
	}); err != nil {
		return err
	}
	if err := fr.setBoolFlag(FlagFeatureStats, func(cgf *types.Config, value bool) {
		fr.cfg.Features.StatsConfig = value
	}); err != nil {
		return err
	}
	if err := fr.setBoolFlag(FlagFeatureClient, func(cgf *types.Config, value bool) {
		fr.cfg.Features.ClientSettings = value
	}); err != nil {
		return err
	}
	if err := fr.setBoolFlag(FlagFeatureServices, func(cgf *types.Config, value bool) {
		fr.cfg.Features.Services = value
	}); err != nil {
		return err
	}
	if err := fr.setBoolFlag(FlagFeatureFilters, func(cgf *types.Config, value bool) {
		fr.cfg.Features.Filters = value
	}); err != nil {
		return err
	}
	return fr.setBoolFlag(FlagFeatureTLSConfig, func(cgf *types.Config, value bool) {
		fr.cfg.Features.TLSConfig = value
	})
}

func (fr *flagReader) readAPIFlags() error {
	if err := fr.setIntFlag(FlagAPIPort, func(cgf *types.Config, value int) {
		fr.cfg.API.Port = value
	}); err != nil {
		return err
	}
	if err := fr.setStringFlag(FlagAPIUsername, func(cgf *types.Config, value string) {
		fr.cfg.API.Username = value
	}); err != nil {
		return err
	}
	if err := fr.setStringFlag(FlagAPIPassword, func(cgf *types.Config, value string) {
		fr.cfg.API.Password = value
	}); err != nil {
		return err
	}
	return fr.setBoolFlag(FlagAPIDarkMode, func(cgf *types.Config, value bool) {
		fr.cfg.API.DarkMode = value
	})
}

func (fr *flagReader) readRootFlags() error {
	if err := fr.setStringFlag(FlagCron, func(cgf *types.Config, value string) {
		fr.cfg.Cron = value
	}); err != nil {
		return err
	}
	if err := fr.setBoolFlag(FlagRunOnStart, func(cgf *types.Config, value bool) {
		fr.cfg.RunOnStart = value
	}); err != nil {
		return err
	}
	if err := fr.setBoolFlag(FlagPrintConfigOnly, func(cgf *types.Config, value bool) {
		fr.cfg.PrintConfigOnly = value
	}); err != nil {
		return err
	}
	return fr.setBoolFlag(FlagContinueOnError, func(cgf *types.Config, value bool) {
		fr.cfg.ContinueOnError = value
	})
}

type Flags interface {
	Changed(name string) bool
	GetString(name string) (string, error)
	GetInt(name string) (int, error)
	GetBool(name string) (bool, error)
}

func (fr *flagReader) setStringFlag(name string, cb callback[string]) (err error) {
	if fr.flags.Changed(name) {
		var value string
		if value, err = fr.flags.GetString(name); err != nil {
			return err
		}
		cb(fr.cfg, value)
	}
	return nil
}

func (fr *flagReader) setBoolFlag(name string, cb callback[bool]) (err error) {
	if fr.flags.Changed(name) {
		var value bool
		if value, err = fr.flags.GetBool(name); err != nil {
			return err
		}
		cb(fr.cfg, value)
	}
	return nil
}

func (fr *flagReader) setIntFlag(name string, cb callback[int]) (err error) {
	if fr.flags.Changed(name) {
		var value int
		if value, err = fr.flags.GetInt(name); err != nil {
			return err
		}
		cb(fr.cfg, value)
	}
	return nil
}

type callback[T any] func(cgf *types.Config, value T)
