package client

import (
	"context"

	"github.com/bakito/adguardhome-sync/pkg/client/model"
	"github.com/bakito/adguardhome-sync/pkg/types"
)

func Run(ai types.AdGuardInstance) error {
	cl, err := New(ai)
	if err != nil {
		return err
	}
	ctx := context.TODO()
	cl.Host(ctx)
	ss, err := cl.GetServerStatus(ctx)
	if err != nil {
		return err
	}
	fc, err := cl.GetFilteringStatus(ctx)
	if err != nil {
		return err
	}
	err = cl.SetFilteringConfig(ctx, model.FilterConfig{Interval: fc.Interval, Enabled: fc.Enabled})
	if err != nil {
		return err
	}
	println(ss)
	return nil
}
