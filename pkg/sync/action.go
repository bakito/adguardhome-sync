package sync

import (
	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/client/model"
	"github.com/bakito/adguardhome-sync/pkg/types"
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
	return
}

type syncAction interface {
	sync(o *origin, client client.Client, rs *model.ServerStatus) error
	name() string
}

type defaultAction struct {
	myName string
	doSync func(o *origin, client client.Client, rs *model.ServerStatus) error
}

func action(name string, f func(o *origin, client client.Client, rs *model.ServerStatus) error) syncAction {
	return &defaultAction{myName: name, doSync: f}
}

func (d *defaultAction) sync(o *origin, client client.Client, rs *model.ServerStatus) error {
	return d.doSync(o, client, rs)
}

func (d *defaultAction) name() string {
	return d.myName
}
