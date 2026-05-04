package config

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	gm "go.uber.org/mock/gomock"

	flagsmock "github.com/bakito/adguardhome-sync/internal/mocks/flags"
	"github.com/bakito/adguardhome-sync/internal/types"
)

func setupFlagsTest(t *testing.T) (*types.Config, *flagsmock.MockFlags, *gm.Controller) {
	t.Helper()
	cfg := &types.Config{
		Origin:  &types.AdGuardInstance{},
		Replica: &types.AdGuardInstance{},
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
			Filters: types.FiltersType{
				Blacklist: true,
				Whitelist: true,
				UserRules:        true,
			},
		},
	}
	mockCtrl := gm.NewController(t)
	flags := flagsmock.NewMockFlags(mockCtrl)
	return cfg, flags, mockCtrl
}

func TestReadFlags_NilFlags(t *testing.T) {
	cfg, _, mockCtrl := setupFlagsTest(t)
	defer mockCtrl.Finish()

	clone := cfg.DeepCopy()
	err := readFlags(cfg, nil)
	if err != nil {
		t.Fatalf("readFlags error = %v, want nil", err)
	}
	if diff := cmp.Diff(clone, cfg); diff != "" {
		t.Errorf("readFlags() mismatch (-want +got):\n%s", diff)
	}
}

func TestReadFlags_NoChangedFlags(t *testing.T) {
	cfg, flags, mockCtrl := setupFlagsTest(t)
	defer mockCtrl.Finish()

	clone := cfg.DeepCopy()
	flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
	err := readFlags(cfg, flags)
	if err != nil {
		t.Fatalf("readFlags error = %v, want nil", err)
	}
	if diff := cmp.Diff(clone, cfg); diff != "" {
		t.Errorf("readFlags() mismatch (-want +got):\n%s", diff)
	}
}

func TestReadFeatureFlags_DisableAll(t *testing.T) {
	cfg, flags, mockCtrl := setupFlagsTest(t)
	defer mockCtrl.Finish()

	flags.EXPECT().Changed(gm.Any()).DoAndReturn(func(name string) bool {
		return strings.HasPrefix(name, "feature")
	}).AnyTimes()
	flags.EXPECT().GetBool(gm.Any()).Return(false, nil).AnyTimes()
	err := readFlags(cfg, flags)
	if err != nil {
		t.Fatalf("readFlags error = %v, want nil", err)
	}

	expectedFeatures := types.Features{
		DNS: types.DNS{
			AccessLists:  false,
			ServerConfig: false,
			Rewrites:     false,
		},
		DHCP: types.DHCP{
			ServerConfig: false,
			StaticLeases: false,
		},
		GeneralSettings: false,
		QueryLogConfig:  false,
		StatsConfig:     false,
		ClientSettings:  false,
		Services:        false,
		Filters: types.FiltersType{
			Blacklist: false,
			Whitelist: false,
			UserRules:        false,
		},
	}

	if diff := cmp.Diff(expectedFeatures, cfg.Features); diff != "" {
		t.Errorf("cfg.Features mismatch (-want +got):\n%s", diff)
	}
}

func TestReadAPIFlags_ChangeAll(t *testing.T) {
	cfg, flags, mockCtrl := setupFlagsTest(t)
	defer mockCtrl.Finish()

	cfg.API = types.API{
		Port:     1111,
		Username: "2222",
		Password: "3333",
		DarkMode: false,
	}
	flags.EXPECT().Changed(gm.Any()).DoAndReturn(func(name string) bool {
		return strings.HasPrefix(name, "api")
	}).AnyTimes()
	flags.EXPECT().GetInt(FlagAPIPort).Return(9999, nil)
	flags.EXPECT().GetString(FlagAPIUsername).Return("aaaa", nil)
	flags.EXPECT().GetString(FlagAPIPassword).Return("bbbb", nil)
	flags.EXPECT().GetBool(FlagAPIDarkMode).Return(true, nil)
	err := readFlags(cfg, flags)
	if err != nil {
		t.Fatalf("readFlags error = %v, want nil", err)
	}

	expectedAPI := types.API{
		Port:     9999,
		Username: "aaaa",
		Password: "bbbb",
		DarkMode: true,
	}

	if diff := cmp.Diff(expectedAPI, cfg.API); diff != "" {
		t.Errorf("cfg.API mismatch (-want +got):\n%s", diff)
	}
}

func TestReadRootFlags_ChangeAll(t *testing.T) {
	cfg, flags, mockCtrl := setupFlagsTest(t)
	defer mockCtrl.Finish()

	cfg.Cron = "*/10 * * * *"
	cfg.PrintConfigOnly = false
	cfg.ContinueOnError = false
	cfg.RunOnStart = false

	flags.EXPECT().Changed(FlagCron).Return(true)
	flags.EXPECT().Changed(FlagRunOnStart).Return(true)
	flags.EXPECT().Changed(FlagPrintConfigOnly).Return(true)
	flags.EXPECT().Changed(FlagContinueOnError).Return(true)
	flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

	flags.EXPECT().GetString(FlagCron).Return("*/30 * * * *", nil)
	flags.EXPECT().GetBool(FlagRunOnStart).Return(true, nil)
	flags.EXPECT().GetBool(FlagPrintConfigOnly).Return(true, nil)
	flags.EXPECT().GetBool(FlagContinueOnError).Return(true, nil)
	err := readFlags(cfg, flags)
	if err != nil {
		t.Fatalf("readFlags error = %v, want nil", err)
	}

	if cfg.Cron != "*/30 * * * *" {
		t.Errorf("cfg.Cron = %s, want */30 * * * *", cfg.Cron)
	}
	if !cfg.RunOnStart {
		t.Error("cfg.RunOnStart = false, want true")
	}
	if !cfg.PrintConfigOnly {
		t.Error("cfg.PrintConfigOnly = false, want true")
	}
	if !cfg.ContinueOnError {
		t.Error("cfg.ContinueOnError = false, want true")
	}
}

func TestReadOriginFlags_ChangeAll(t *testing.T) {
	cfg, flags, mockCtrl := setupFlagsTest(t)
	defer mockCtrl.Finish()

	cfg.Origin = &types.AdGuardInstance{
		URL:                "1",
		WebURL:             "2",
		APIPath:            "3",
		Username:           "4",
		Password:           "5",
		Cookie:             "6",
		InsecureSkipVerify: false,
	}

	flags.EXPECT().Changed(FlagOriginURL).Return(true)
	flags.EXPECT().Changed(FlagOriginWebURL).Return(true)
	flags.EXPECT().Changed(FlagOriginAPIPath).Return(true)
	flags.EXPECT().Changed(FlagOriginUsername).Return(true)
	flags.EXPECT().Changed(FlagOriginPassword).Return(true)
	flags.EXPECT().Changed(FlagOriginCookie).Return(true)
	flags.EXPECT().Changed(FlagOriginISV).Return(true)
	flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

	flags.EXPECT().GetString(FlagOriginURL).Return("a", nil)
	flags.EXPECT().GetString(FlagOriginWebURL).Return("b", nil)
	flags.EXPECT().GetString(FlagOriginAPIPath).Return("c", nil)
	flags.EXPECT().GetString(FlagOriginUsername).Return("d", nil)
	flags.EXPECT().GetString(FlagOriginPassword).Return("e", nil)
	flags.EXPECT().GetString(FlagOriginCookie).Return("f", nil)
	flags.EXPECT().GetBool(FlagOriginISV).Return(true, nil)
	err := readFlags(cfg, flags)
	if err != nil {
		t.Fatalf("readFlags error = %v, want nil", err)
	}

	expectedOrigin := &types.AdGuardInstance{
		URL:                "a",
		WebURL:             "b",
		APIPath:            "c",
		Username:           "d",
		Password:           "e",
		Cookie:             "f",
		InsecureSkipVerify: true,
	}

	if diff := cmp.Diff(expectedOrigin, cfg.Origin); diff != "" {
		t.Errorf("cfg.Origin mismatch (-want +got):\n%s", diff)
	}
}

func TestReadReplicaFlags_ChangeAll(t *testing.T) {
	cfg, flags, mockCtrl := setupFlagsTest(t)
	defer mockCtrl.Finish()

	cfg.Replica = &types.AdGuardInstance{
		URL:                "1",
		WebURL:             "2",
		APIPath:            "3",
		Username:           "4",
		Password:           "5",
		Cookie:             "6",
		InsecureSkipVerify: false,
		AutoSetup:          false,
		InterfaceName:      "7",
	}

	flags.EXPECT().Changed(FlagReplicaURL).Return(true)
	flags.EXPECT().Changed(FlagReplicaWebURL).Return(true)
	flags.EXPECT().Changed(FlagReplicaAPIPath).Return(true)
	flags.EXPECT().Changed(FlagReplicaUsername).Return(true)
	flags.EXPECT().Changed(FlagReplicaPassword).Return(true)
	flags.EXPECT().Changed(FlagReplicaCookie).Return(true)
	flags.EXPECT().Changed(FlagReplicaISV).Return(true)
	flags.EXPECT().Changed(FlagReplicaAutoSetup).Return(true)
	flags.EXPECT().Changed(FlagReplicaInterfaceName).Return(true)
	flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

	flags.EXPECT().GetString(FlagReplicaURL).Return("a", nil)
	flags.EXPECT().GetString(FlagReplicaWebURL).Return("b", nil)
	flags.EXPECT().GetString(FlagReplicaAPIPath).Return("c", nil)
	flags.EXPECT().GetString(FlagReplicaUsername).Return("d", nil)
	flags.EXPECT().GetString(FlagReplicaPassword).Return("e", nil)
	flags.EXPECT().GetString(FlagReplicaCookie).Return("f", nil)
	flags.EXPECT().GetBool(FlagReplicaISV).Return(true, nil)
	flags.EXPECT().GetBool(FlagReplicaAutoSetup).Return(true, nil)
	flags.EXPECT().GetString(FlagReplicaInterfaceName).Return("g", nil)
	err := readFlags(cfg, flags)
	if err != nil {
		t.Fatalf("readFlags error = %v, want nil", err)
	}

	expectedReplica := &types.AdGuardInstance{
		URL:                "a",
		WebURL:             "b",
		APIPath:            "c",
		Username:           "d",
		Password:           "e",
		Cookie:             "f",
		InsecureSkipVerify: true,
		AutoSetup:          true,
		InterfaceName:      "g",
	}

	if diff := cmp.Diff(expectedReplica, cfg.Replica); diff != "" {
		t.Errorf("cfg.Replica mismatch (-want +got):\n%s", diff)
	}
}
