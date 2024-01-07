package sync

import (
	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/client/model"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"go.uber.org/zap"
)

func setupActions(cfg *types.Config) (actions []syncAction) {
	if cfg.Features.GeneralSettings {
		actions = append(actions,
			action("profile info", actionProfileInfo),
			action("protection", actionProtection),
			action("parental", actionParental),
			action("safe search config", actionSafeSearchConfig),
			action("safe browsing", actionSafeBrowsing),
		)
	}
	if cfg.Features.QueryLogConfig {
		actions = append(actions,
			action("query log config", actionQueryLogConfig),
		)
	}
	if cfg.Features.StatsConfig {
		actions = append(actions,
			action("stats config", actionStatsConfig),
		)
	}
	if cfg.Features.DNS.Rewrites {
		actions = append(actions,
			action("DNS rewrites", dnsRewrites),
		)
	}
	if cfg.Features.Filters {
		actions = append(actions,
			action("filters", filters),
		)
	}
	if cfg.Features.Services {
		actions = append(actions,
			action("blocked services", blockedServices),
			action("blocked services schedule", blockedServicesSchedule),
		)
	}
	if cfg.Features.ClientSettings {
		actions = append(actions,
			action("client settings", clientSettings),
		)
	}
	if cfg.Features.DNS.AccessLists {
		actions = append(actions,
			action("DNS access lists", dnsAccessLists),
		)
	}

	if cfg.Features.DNS.ServerConfig {
		actions = append(actions,
			action("DNS server config", dnsServerConfig),
		)
	}
	if cfg.Features.DHCP.ServerConfig {
		actions = append(actions,
			action("DHCP server config", dhcpServerConfig),
		)
	}
	if cfg.Features.DHCP.StaticLeases {
		actions = append(actions,
			action("DHCP static leases", dhcpStaticLeases),
		)
	}
	return
}

type syncAction interface {
	sync(ac *actionContext) error
	name() string
}

type actionContext struct {
	rl              *zap.SugaredLogger
	o               *origin
	client          client.Client
	rs              *model.ServerStatus
	continueOnError bool
	replica         types.AdGuardInstance
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
