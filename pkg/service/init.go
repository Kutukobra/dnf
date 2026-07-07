package service

import (
	"context"

	"github.com/free5gc/dnf/internal/logger"
	"github.com/free5gc/dnf/pkg/factory"
)

var DNF *DnfApp

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

func (a *DnfApp) Start() {
	logger.InitLog.Infoln("Server started")
}
