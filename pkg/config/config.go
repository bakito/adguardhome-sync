package config

import (
	"errors"
	"regexp"

	"github.com/caarlos0/env/v11"

	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
)

var (
	envReplicasURLPattern = regexp.MustCompile(`^REPLICA(\d+)_URL=(.*)`)
	logger                = log.GetLogger("config")
)

type AppConfig struct {
	cfg      *types.Config
	filePath string
	content  string
}

func (ac *AppConfig) PrintConfigOnly() bool {
	return ac.cfg.PrintConfigOnly
}

func (ac *AppConfig) Get() *types.Config {
	return ac.cfg
}

func (ac *AppConfig) Init() error {
	return ac.cfg.Init()
}

func Get(configFile string, flags Flags) (*AppConfig, error) {
	path, err := configFilePath(configFile)
	if err != nil {
		return nil, err
	}

	if err := validateSchema(path); err != nil {
		return nil, err
	}

	cfg := initialConfig()

	// read yaml config
	var content string
	if content, err = readFile(cfg, path); err != nil {
		return nil, err
	}

	// overwrite from command flags
	if err := readFlags(cfg, flags); err != nil {
		return nil, err
	}

	// *bool field creates issues when already not nil
	cfg.Origin.DHCPServerEnabled = nil // origin filed makes no sense to be set.

	// keep previously set value
	replicaDhcpServer := cfg.Replica.DHCPServerEnabled
	cfg.Replica.DHCPServerEnabled = nil

	// ignore replicas form env parsing as they are handled separately
	replicas := cfg.Replicas
	cfg.Replicas = nil

	// overwrite from env vars
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	if err := env.ParseWithOptions(cfg.Replica, env.Options{Prefix: "REPLICA_"}); err != nil {
		return nil, err
	}
	// restore the replica
	cfg.Replicas = replicas

	// if not set from env, use previous value
	if cfg.Replica.DHCPServerEnabled == nil {
		cfg.Replica.DHCPServerEnabled = replicaDhcpServer
	}

	if err := env.ParseWithOptions(&cfg.Origin, env.Options{Prefix: "ORIGIN_"}); err != nil {
		return nil, err
	}

	if cfg.Replica != nil &&
		cfg.Replica.URL == "" &&
		cfg.Replica.Username == "" {
		cfg.Replica = nil
	}

	if len(cfg.Replicas) > 0 && cfg.Replica != nil {
		return nil, errors.New("mixed replica config in use. " +
			"Do not use single replica and numbered (list) replica config combined " + cfg.Replica.Username)
	}

	handleDeprecatedEnvVars(cfg)

	if cfg.Replica != nil {
		cfg.Replicas = []types.AdGuardInstance{*cfg.Replica}
		cfg.Replica = nil
	}

	cfg.Replicas, err = enrichReplicasFromEnv(cfg.Replicas)

	return &AppConfig{cfg: cfg, filePath: path, content: content}, err
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
		Features: types.NewFeatures(true),
	}
}
