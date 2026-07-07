package service

import (
	"context"

	"github.com/free5gc/dnf/pkg/factory"
)

type DnfApp struct {
	cfg *factory.Config

	ctx context.Context
}

func NewApp(ctx context.Context, cfg *factory.Config) (*DnfApp, error) {
	dnf := &DnfApp{
		cfg: cfg,
		ctx: ctx,
	}

	return dnf, nil
}
