package config_test

import (
	"strings"
	"testing"

	gm "go.uber.org/mock/gomock"

	"github.com/bakito/adguardhome-sync/internal/config"
	flagsmock "github.com/bakito/adguardhome-sync/internal/mocks/flags"
)

type configTestHelper struct {
	flags    *flagsmock.MockFlags
	mockCtrl *gm.Controller
}

func newConfigTestHelper(t *testing.T) *configTestHelper {
	t.Helper()
	mockCtrl := gm.NewController(t)
	return &configTestHelper{
		mockCtrl: mockCtrl,
		flags:    flagsmock.NewMockFlags(mockCtrl),
	}
}

func (*configTestHelper) setEnv(t *testing.T, name, value string) {
	t.Helper()
	t.Setenv(name, value)
}

func (h *configTestHelper) finish() {
	h.mockCtrl.Finish()
}

func TestConfigGet_MixedConfig(t *testing.T) {
	h := newConfigTestHelper(t)
	defer h.finish()

	h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

	_, err := config.Get("../../testdata/config_test_replicas_and_replica.yaml", h.flags)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "mixed replica config in use") {
		t.Errorf("expected error containing 'mixed replica config in use', got '%s'", err.Error())
	}
}

func TestConfigGet_EnvVarClash(t *testing.T) {
	h := newConfigTestHelper(t)
	defer h.finish()

	incorrect := "ThisIsNotTheCorrectUsername"
	h.setEnv(t, "USERNAME", incorrect)
	h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

	c, err := config.Get("../../testdata/config_test_replica.yaml", h.flags)
	if err != nil {
		t.Fatalf("config.Get error = %v, want nil", err)
	}
	if c.Get().Origin.Username == incorrect {
		t.Errorf("origin username should not be %s", incorrect)
	}
	if c.Get().Replicas[0].Username == incorrect {
		t.Errorf("replica username should not be %s", incorrect)
	}
}

func TestConfigGet_OriginUrl(t *testing.T) {
	t.Run("from config file", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().Origin.URL != "https://origin-file:443" {
			t.Errorf("origin URL = %s, want https://origin-file:443", cfg.Get().Origin.URL)
		}
	})

	t.Run("from config flags", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(config.FlagOriginURL).Return(true).AnyTimes()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		h.flags.EXPECT().GetString(config.FlagOriginURL).Return("https://origin-flag:443", nil).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().Origin.URL != "https://origin-flag:443" {
			t.Errorf("origin URL = %s, want https://origin-flag:443", cfg.Get().Origin.URL)
		}
	})

	t.Run("from config env var", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.setEnv(t, "ORIGIN_URL", "https://origin-env:443")
		h.flags.EXPECT().Changed(config.FlagOriginURL).Return(true).AnyTimes()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		h.flags.EXPECT().GetString(config.FlagOriginURL).Return("https://origin-flag:443", nil).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().Origin.URL != "https://origin-env:443" {
			t.Errorf("origin URL = %s, want https://origin-env:443", cfg.Get().Origin.URL)
		}
	})
}

func TestConfigGet_ReplicaInsecureSkipVerify(t *testing.T) {
	t.Run("from config file", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replica.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().Replicas[0].InsecureSkipVerify {
			t.Error("replica InsecureSkipVerify = true, want false")
		}
	})

	t.Run("from config flags", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(config.FlagReplicaISV).Return(true).AnyTimes()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		h.flags.EXPECT().GetBool(config.FlagReplicaISV).Return(true, nil).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replica.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if !cfg.Get().Replicas[0].InsecureSkipVerify {
			t.Error("replica InsecureSkipVerify = false, want true")
		}
	})

	t.Run("from config env var", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.setEnv(t, "REPLICA_INSECURE_SKIP_VERIFY", "false")
		h.flags.EXPECT().Changed(config.FlagReplicaISV).Return(true).AnyTimes()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		h.flags.EXPECT().GetBool(config.FlagReplicaISV).Return(true, nil).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replica.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().Replicas[0].InsecureSkipVerify {
			t.Error("replica InsecureSkipVerify = true, want false")
		}
	})
}

func TestConfigGet_Replica1InsecureSkipVerify(t *testing.T) {
	t.Run("from config file", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().Replicas[0].InsecureSkipVerify {
			t.Error("replica InsecureSkipVerify = true, want false")
		}
	})

	t.Run("from config env var", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.setEnv(t, "REPLICA1_INSECURE_SKIP_VERIFY", "true")
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if !cfg.Get().Replicas[0].InsecureSkipVerify {
			t.Error("replica InsecureSkipVerify = false, want true")
		}
	})
}

func TestConfigGet_APIPort(t *testing.T) {
	t.Run("from config file", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().API.Port != 9090 {
			t.Errorf("API Port = %d, want 9090", cfg.Get().API.Port)
		}
	})

	t.Run("from config flags", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(config.FlagAPIPort).Return(true).AnyTimes()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		h.flags.EXPECT().GetInt(config.FlagAPIPort).Return(9990, nil).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().API.Port != 9990 {
			t.Errorf("API Port = %d, want 9990", cfg.Get().API.Port)
		}
	})

	t.Run("from config env var", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.setEnv(t, "API_PORT", "9999")
		h.flags.EXPECT().Changed(config.FlagAPIPort).Return(true).AnyTimes()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		h.flags.EXPECT().GetInt(config.FlagAPIPort).Return(9990, nil).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().API.Port != 9999 {
			t.Errorf("API Port = %d, want 9999", cfg.Get().API.Port)
		}
	})
}

func TestConfigGet_ReplicaDHCPServerEnabled(t *testing.T) {
	h := newConfigTestHelper(t)
	defer h.finish()
	h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

	cfg, err := config.Get("../../testdata/config_test_replica.yaml", h.flags)
	if err != nil {
		t.Fatalf("config.Get error = %v, want nil", err)
	}
	if cfg.Get().Replicas[0].DHCPServerEnabled == nil {
		t.Fatal("replica DHCPServerEnabled is nil")
	}
	if *cfg.Get().Replicas[0].DHCPServerEnabled {
		t.Error("replica DHCPServerEnabled = true, want false")
	}
}

func TestConfigGet_Replica1DHCPServerEnabled(t *testing.T) {
	h := newConfigTestHelper(t)
	defer h.finish()
	h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

	cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
	if err != nil {
		t.Fatalf("config.Get error = %v, want nil", err)
	}
	if cfg.Get().Replicas[0].DHCPServerEnabled == nil {
		t.Fatal("replica DHCPServerEnabled is nil")
	}
	if *cfg.Get().Replicas[0].DHCPServerEnabled {
		t.Error("replica DHCPServerEnabled = true, want false")
	}
}

func TestConfigGet_FeatureDNSServerConfig(t *testing.T) {
	t.Run("from config file", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().Features.DNS.ServerConfig {
			t.Error("feature DNS server config = true, want false")
		}
	})

	t.Run("from config flags", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(config.FlagFeatureDNSServerConfig).Return(true).AnyTimes()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		h.flags.EXPECT().GetBool(config.FlagFeatureDNSServerConfig).Return(true, nil).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if !cfg.Get().Features.DNS.ServerConfig {
			t.Error("feature DNS server config = false, want true")
		}
	})

	t.Run("from config env var", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.setEnv(t, "FEATURES_DNS_SERVER_CONFIG", "false")
		h.flags.EXPECT().Changed(config.FlagFeatureDNSServerConfig).Return(true).AnyTimes()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		h.flags.EXPECT().GetBool(config.FlagFeatureDNSServerConfig).Return(true, nil).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().Features.DNS.ServerConfig {
			t.Error("feature DNS server config = true, want false")
		}
	})
}

func TestConfigGet_FeatureFiltersBlacklist(t *testing.T) {
	t.Run("from config file", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if !cfg.Get().Features.Filters.Blacklist {
			t.Error("feature filters blacklist = false, want true")
		}
	})

	t.Run("from config flags", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(config.FlagFeatureFiltersBlacklist).Return(true).AnyTimes()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		h.flags.EXPECT().GetBool(config.FlagFeatureFiltersBlacklist).Return(false, nil).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().Features.Filters.Blacklist {
			t.Error("feature filters blacklist = true, want false")
		}
	})

	t.Run("from config env var", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.setEnv(t, "FEATURES_FILTERS_BLACKLIST", "false")
		h.flags.EXPECT().Changed(config.FlagFeatureFiltersBlacklist).Return(true).AnyTimes()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		h.flags.EXPECT().GetBool(config.FlagFeatureFiltersBlacklist).Return(true, nil).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().Features.Filters.Blacklist {
			t.Error("feature filters blacklist = true, want false")
		}
	})
}

func TestConfigGet_FeatureFiltersWhitelist(t *testing.T) {
	t.Run("from config file", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().Features.Filters.Whitelist {
			t.Error("feature filters whitelist = true, want false")
		}
	})

	t.Run("from config flags", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(config.FlagFeatureFiltersWhitelist).Return(true).AnyTimes()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		h.flags.EXPECT().GetBool(config.FlagFeatureFiltersWhitelist).Return(true, nil).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if !cfg.Get().Features.Filters.Whitelist {
			t.Error("feature filters whitelist = false, want true")
		}
	})

	t.Run("from config env var", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.setEnv(t, "FEATURES_FILTERS_WHITELIST", "false")
		h.flags.EXPECT().Changed(config.FlagFeatureFiltersWhitelist).Return(true).AnyTimes()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		h.flags.EXPECT().GetBool(config.FlagFeatureFiltersWhitelist).Return(true, nil).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().Features.Filters.Whitelist {
			t.Error("feature filters whitelist = true, want false")
		}
	})
}

func TestConfigGet_FeatureFiltersUserRules(t *testing.T) {
	t.Run("from config file", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if !cfg.Get().Features.Filters.UserRules {
			t.Error("feature filters user rules = false, want true")
		}
	})

	t.Run("from config flags", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(config.FlagFeatureFiltersUserRules).Return(true).AnyTimes()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		h.flags.EXPECT().GetBool(config.FlagFeatureFiltersUserRules).Return(false, nil).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().Features.Filters.UserRules {
			t.Error("feature filters user rules = true, want false")
		}
	})

	t.Run("from config env var", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.setEnv(t, "FEATURES_FILTERS_USER_RULES", "false")
		h.flags.EXPECT().Changed(config.FlagFeatureFiltersUserRules).Return(true).AnyTimes()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
		h.flags.EXPECT().GetBool(config.FlagFeatureFiltersUserRules).Return(true, nil).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if cfg.Get().Features.Filters.UserRules {
			t.Error("feature filters user rules = true, want false")
		}
	})
}

func TestConfigGet_Headers(t *testing.T) {
	t.Run("from config file", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if len(cfg.Get().Replicas[0].RequestHeaders) != 2 {
			t.Errorf("headers length = %d, want 2", len(cfg.Get().Replicas[0].RequestHeaders))
		}
		if cfg.Get().Replicas[0].RequestHeaders["FOO"] != "bar" {
			t.Errorf("FOO header = %s, want bar", cfg.Get().Replicas[0].RequestHeaders["FOO"])
		}
		if cfg.Get().Replicas[0].RequestHeaders["Client-ID"] != "xxxx" {
			t.Errorf("Client-ID header = %s, want xxxx", cfg.Get().Replicas[0].RequestHeaders["Client-ID"])
		}
	})

	t.Run("from config file replaced by ENV", func(t *testing.T) {
		h := newConfigTestHelper(t)
		defer h.finish()
		h.setEnv(t, "REPLICA1_REQUEST_HEADERS", "AAA:bbb")
		h.flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

		cfg, err := config.Get("../../testdata/config_test_replicas.yaml", h.flags)
		if err != nil {
			t.Fatalf("config.Get error = %v, want nil", err)
		}
		if len(cfg.Get().Replicas[0].RequestHeaders) != 1 {
			t.Errorf("headers length = %d, want 1", len(cfg.Get().Replicas[0].RequestHeaders))
		}
		if cfg.Get().Replicas[0].RequestHeaders["AAA"] != "bbb" {
			t.Errorf("AAA header = %s, want bbb", cfg.Get().Replicas[0].RequestHeaders["AAA"])
		}
	})
}
