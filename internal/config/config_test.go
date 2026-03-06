package config_test

import (
	"strings"
	"testing"

	gm "go.uber.org/mock/gomock"

	"github.com/bakito/adguardhome-sync/internal/config"
	flagsmock "github.com/bakito/adguardhome-sync/internal/mocks/flags"
)

func TestGet(t *testing.T) {
	setup := func(t *testing.T) (*flagsmock.MockFlags, func(string, string)) {
		t.Helper()
		mockCtrl := gm.NewController(t)
		flags := flagsmock.NewMockFlags(mockCtrl)
		var changedEnvVars []string
		setEnv := func(name, value string) {
			t.Setenv(name, value)
			changedEnvVars = append(changedEnvVars, name)
		}
		t.Cleanup(func() {
			mockCtrl.Finish()
		})
		return flags, setEnv
	}

	t.Run("Mixed Config", func(t *testing.T) {
		flags, _ := setup(t)
		flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

		_, err := config.Get("../../testdata/config_test_replicas_and_replica.yaml", flags)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		if !strings.Contains(err.Error(), "mixed replica config in use") {
			t.Errorf("expected error to contain 'mixed replica config in use' but got '%v'", err)
		}
	})

	t.Run("Env Var Clash", func(t *testing.T) {
		flags, setEnv := setup(t)
		incorrect := "ThisIsNotTheCorrectUsername"
		setEnv("USERNAME", incorrect)
		flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

		c, err := config.Get("../../testdata/config_test_replica.yaml", flags)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Get().Origin.Username == incorrect {
			t.Errorf("origin username should not be '%s'", incorrect)
		}
		if c.Get().Replicas[0].Username == incorrect {
			t.Errorf("replica username should not be '%s'", incorrect)
		}
	})

	t.Run("Origin Url", func(t *testing.T) {
		t.Run("should have the origin URL from the config file", func(t *testing.T) {
			flags, _ := setup(t)
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Get().Origin.URL != "https://origin-file:443" {
				t.Errorf("expected https://origin-file:443 but got %s", cfg.Get().Origin.URL)
			}
		})
		t.Run("should have the origin URL from the config flags", func(t *testing.T) {
			flags, _ := setup(t)
			flags.EXPECT().Changed(config.FlagOriginURL).Return(true).AnyTimes()
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
			flags.EXPECT().GetString(config.FlagOriginURL).Return("https://origin-flag:443", nil).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Get().Origin.URL != "https://origin-flag:443" {
				t.Errorf("expected https://origin-flag:443 but got %s", cfg.Get().Origin.URL)
			}
		})
		t.Run("should have the origin URL from the config env var", func(t *testing.T) {
			flags, setEnv := setup(t)
			setEnv("ORIGIN_URL", "https://origin-env:443")
			flags.EXPECT().Changed(config.FlagOriginURL).Return(true).AnyTimes()
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
			flags.EXPECT().GetString(config.FlagOriginURL).Return("https://origin-flag:443", nil).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Get().Origin.URL != "https://origin-env:443" {
				t.Errorf("expected https://origin-env:443 but got %s", cfg.Get().Origin.URL)
			}
		})
	})

	t.Run("Replica insecure skip verify", func(t *testing.T) {
		t.Run("should have the insecure skip verify from the config file", func(t *testing.T) {
			flags, _ := setup(t)
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replica.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Get().Replicas[0].InsecureSkipVerify {
				t.Error("expected false but got true")
			}
		})
		t.Run("should have the insecure skip verify from the config flags", func(t *testing.T) {
			flags, _ := setup(t)
			flags.EXPECT().Changed(config.FlagReplicaISV).Return(true).AnyTimes()
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
			flags.EXPECT().GetBool(config.FlagReplicaISV).Return(true, nil).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replica.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !cfg.Get().Replicas[0].InsecureSkipVerify {
				t.Error("expected true but got false")
			}
		})
		t.Run("should have the insecure skip verify from the config env var", func(t *testing.T) {
			flags, setEnv := setup(t)
			setEnv("REPLICA_INSECURE_SKIP_VERIFY", "false")
			flags.EXPECT().Changed(config.FlagReplicaISV).Return(true).AnyTimes()
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
			flags.EXPECT().GetBool(config.FlagReplicaISV).Return(true, nil).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replica.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Get().Replicas[0].InsecureSkipVerify {
				t.Error("expected false but got true")
			}
		})
	})

	t.Run("Replica 1 insecure skip verify", func(t *testing.T) {
		t.Run("should have the insecure skip verify from the config file", func(t *testing.T) {
			flags, _ := setup(t)
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Get().Replicas[0].InsecureSkipVerify {
				t.Error("expected false but got true")
			}
		})
		t.Run("should have the insecure skip verify from the config env var", func(t *testing.T) {
			flags, setEnv := setup(t)
			setEnv("REPLICA1_INSECURE_SKIP_VERIFY", "true")
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !cfg.Get().Replicas[0].InsecureSkipVerify {
				t.Error("expected true but got false")
			}
		})
	})

	t.Run("API Port", func(t *testing.T) {
		t.Run("should have the api port from the config file", func(t *testing.T) {
			flags, _ := setup(t)
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
			cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Get().API.Port != 9090 {
				t.Errorf("expected 9090 but got %d", cfg.Get().API.Port)
			}
		})
		t.Run("should have the api port from the config flags", func(t *testing.T) {
			flags, _ := setup(t)
			flags.EXPECT().Changed(config.FlagAPIPort).Return(true).AnyTimes()
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
			flags.EXPECT().GetInt(config.FlagAPIPort).Return(9990, nil).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Get().API.Port != 9990 {
				t.Errorf("expected 9990 but got %d", cfg.Get().API.Port)
			}
		})
		t.Run("should have the api port from the config env var", func(t *testing.T) {
			flags, setEnv := setup(t)
			setEnv("API_PORT", "9999")
			flags.EXPECT().Changed(config.FlagAPIPort).Return(true).AnyTimes()
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
			flags.EXPECT().GetInt(config.FlagAPIPort).Return(9990, nil).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Get().API.Port != 9999 {
				t.Errorf("expected 9999 but got %d", cfg.Get().API.Port)
			}
		})
	})

	t.Run("Replica DHCPServerEnabled", func(t *testing.T) {
		t.Run("should have the dhcp server enabled from the config file", func(t *testing.T) {
			flags, _ := setup(t)
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replica.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Get().Replicas[0].DHCPServerEnabled == nil {
				t.Fatal("expected DHCPServerEnabled to be non-nil")
			}
			if *cfg.Get().Replicas[0].DHCPServerEnabled {
				t.Errorf("expected false but got true")
			}
		})
	})

	t.Run("Replica 1 DHCPServerEnabled", func(t *testing.T) {
		t.Run("should have the dhcp server enabled from the config file", func(t *testing.T) {
			flags, _ := setup(t)
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Get().Replicas[0].DHCPServerEnabled == nil {
				t.Fatal("expected DHCPServerEnabled to be non-nil")
			}
			if *cfg.Get().Replicas[0].DHCPServerEnabled {
				t.Errorf("expected false but got true")
			}
		})
	})

	t.Run("Feature DNS Server Config", func(t *testing.T) {
		t.Run("should have the feature dns server config from the config file", func(t *testing.T) {
			flags, _ := setup(t)
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
			cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Get().Features.DNS.ServerConfig {
				t.Errorf("expected false but got true")
			}
		})
		t.Run("should have the feature dns server config from the config flags", func(t *testing.T) {
			flags, _ := setup(t)
			flags.EXPECT().Changed(config.FlagFeatureDNSServerConfig).Return(true).AnyTimes()
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
			flags.EXPECT().GetBool(config.FlagFeatureDNSServerConfig).Return(true, nil).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !cfg.Get().Features.DNS.ServerConfig {
				t.Errorf("expected true but got false")
			}
		})
		t.Run("should have the feature dns server config from the config env var", func(t *testing.T) {
			flags, setEnv := setup(t)
			setEnv("FEATURES_DNS_SERVER_CONFIG", "false")
			flags.EXPECT().Changed(config.FlagFeatureDNSServerConfig).Return(true).AnyTimes()
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
			flags.EXPECT().GetBool(config.FlagFeatureDNSServerConfig).Return(true, nil).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Get().Features.DNS.ServerConfig {
				t.Errorf("expected false but got true")
			}
		})
	})

	t.Run("Headers", func(t *testing.T) {
		t.Run("have headers from the config file", func(t *testing.T) {
			flags, _ := setup(t)
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(cfg.Get().Replicas[0].RequestHeaders) != 2 {
				t.Errorf("expected 2 headers but got %d", len(cfg.Get().Replicas[0].RequestHeaders))
			}
			if cfg.Get().Replicas[0].RequestHeaders["FOO"] != "bar" {
				t.Errorf("expected bar but got %s", cfg.Get().Replicas[0].RequestHeaders["FOO"])
			}
			if cfg.Get().Replicas[0].RequestHeaders["Client-ID"] != "xxxx" {
				t.Errorf("expected xxxx but got %s", cfg.Get().Replicas[0].RequestHeaders["Client-ID"])
			}
		})
		t.Run("have headers from the config file will be replaced when defined as ENV", func(t *testing.T) {
			flags, setEnv := setup(t)
			setEnv("REPLICA1_REQUEST_HEADERS", "AAA:bbb")
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

			cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(cfg.Get().Replicas[0].RequestHeaders) != 1 {
				t.Errorf("expected 1 header but got %d", len(cfg.Get().Replicas[0].RequestHeaders))
			}
			if cfg.Get().Replicas[0].RequestHeaders["AAA"] != "bbb" {
				t.Errorf("expected bbb but got %s", cfg.Get().Replicas[0].RequestHeaders["AAA"])
			}
		})
	})
}
