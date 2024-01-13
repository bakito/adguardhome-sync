package config

import (
	"errors"
	"regexp"

	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/caarlos0/env/v10"
	"github.com/spf13/cobra"
)

var (
	envReplicasURLPattern = regexp.MustCompile(`^REPLICA(\d+)_URL=(.*)`)
	logger                = log.GetLogger("config")
)

func Get(configFile string, cmd *cobra.Command) (*types.Config, error) {
	path, err := configFilePath(configFile)
	if err != nil {
		return nil, err
	}

	cfg := &types.Config{
		Replica: &types.AdGuardInstance{},
	}

	// read yaml config
	if err := readFile(cfg, path); err != nil {
		return nil, err
	}

	// overwrite from command flags
	if err := readFlags(cfg, cmd); err != nil {
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
