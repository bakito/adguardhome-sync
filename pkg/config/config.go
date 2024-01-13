package config

import (
	"errors"
	"regexp"

	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/caarlos0/env/v10"
)

var (
	envReplicasURLPattern = regexp.MustCompile(`^REPLICA(\d+)_URL=(.*)`)
	logger                = log.GetLogger("config")
)

func Get(configFile string, flags Flags) (*types.Config, error) {
	path, err := configFilePath(configFile)
	if err != nil {
		return nil, err
	}

	cfg := initialConfig()

	// read yaml config
	if err := readFile(cfg, path); err != nil {
		return nil, err
	}

	// overwrite from command flags
	if err := readFlags(cfg, flags); err != nil {
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

	cfg.Replicas, err = enrichReplicasFromEnv(cfg.Replicas)

	return cfg, err
}

func initialConfig() *types.Config {
	return &types.Config{
		RunOnStart: true,
		Origin: types.AdGuardInstance{
			APIPath: "/control",
		},
		Replica: &types.AdGuardInstance{
			APIPath: "/control",
		},
		API: types.API{
			Port: 8080,
		},
		Features: types.Features{
			DNS: types.DNS{
				AccessLists:  true,
				ServerConfig: true,
				Rewrites:     true,
			},
			DHCP: types.DHCP{
				ServerConfig: true,
				StaticLeases: true,
			},
			GeneralSettings: true,
			QueryLogConfig:  true,
			StatsConfig:     true,
			ClientSettings:  true,
			Services:        true,
			Filters:         true,
		},
	}
}
