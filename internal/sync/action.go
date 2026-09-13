package sync

import (
	"go.uber.org/zap"

	"github.com/bakito/adguardhome-sync/internal/client"
	"github.com/bakito/adguardhome-sync/internal/client/model"
	"github.com/bakito/adguardhome-sync/internal/types"
)

func setupActions(features types.Features) (actions []syncAction) {
	if features.GeneralSettings {
		actions = append(actions, action("profile info", actionProfileInfo))
		if features.ProtectionStatus {
			actions = append(actions, action("protection", actionProtection))
		}
		actions = append(actions,
			action("parental", actionParental),
			action("safe search config", actionSafeSearchConfig),
			action("safe browsing", actionSafeBrowsing),
		)
	}
	if features.DNS.ServerConfig {
		actions = append(actions,
			action("DNS server config", actionDNSServerConfig),
		)
	}
	if features.QueryLogConfig {
		actions = append(actions,
			action("query log config", actionQueryLogConfig),
		)
	}
	if features.StatsConfig {
		actions = append(actions,
			action("stats config", actionStatsConfig),
		)
	}
	if features.DNS.Rewrites {
		actions = append(actions,
			action("DNS rewrite settings", actionRewriteSettings),
			action("DNS rewrite entries", actionRewriteEntries),
		)
	}
	if features.Filters.Blacklist || features.Filters.Whitelist || features.Filters.UserRules {
		actions = append(actions,
			action("actionFilters", actionFilters),
		)
	}
	if features.Services {
		actions = append(actions,
			action("blocked services schedule", actionBlockedServicesSchedule),
		)
	}
	if features.ClientSettings {
		actions = append(actions,
			action("client settings", actionClientSettings),
		)
	}
	if features.DNS.AccessLists {
		actions = append(actions,
			action("DNS access lists", actionDNSAccessLists),
		)
	}
	if features.DHCP.ServerConfig {
		actions = append(actions,
			action("DHCP server config", actionDHCPServerConfig),
		)
	}
	if features.DHCP.StaticLeases {
		actions = append(actions,
			action("DHCP static leases", actionDHCPStaticLeases),
		)
	}
	if features.TLSConfig {
		actions = append(actions,
			action("TLS config", tlsConfig),
		)
	}
	return actions
}

type syncAction interface {
	sync(ac *actionContext) error
	name() string
}

type actionContext struct {
	rl            *zap.SugaredLogger
	origin        *origin
	client        client.Client
	replicaStatus *model.ServerStatus
	replica       types.Replica
	cfg           *types.Config
}

func (ac *actionContext) features() types.Features {
	return ac.replica.EffectiveFeatures(ac.cfg.Features)
}

type defaultAction struct {
	myName string
	doSync func(ac *actionContext) error
}

func action(name string, f func(ac *actionContext) error) syncAction {
	return &defaultAction{myName: name, doSync: f}
}

func (d *defaultAction) sync(ac *actionContext) error {
	return d.doSync(ac)
}

func (d *defaultAction) name() string {
	return d.myName
}
