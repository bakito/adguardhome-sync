package client

import (
	"context"

	"github.com/bakito/adguardhome-sync/internal/client/model"
)

type Client interface {
	Host(ctx context.Context) string
	GetServerStatus(ctx context.Context) (*model.ServerStatus, error)

	GetFilteringStatus(ctx context.Context) (*model.FilterStatus, error)
	SetFilteringConfig(ctx context.Context, config model.FilterConfig) error
}
