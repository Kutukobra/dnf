package service

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"

	dnf_context "github.com/free5gc/dnf/internal/context"
	"github.com/free5gc/dnf/internal/logger"
	"github.com/free5gc/dnf/internal/sbi"
	"github.com/free5gc/dnf/internal/sbi/consumer"
	"github.com/free5gc/dnf/internal/sbi/processor"
	"github.com/free5gc/dnf/pkg/factory"
)

var DNF *DnfApp

type DnfApp struct {
	dnfCtx *dnf_context.DnfContext
	cfg    *factory.Config

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	sbiServer *sbi.Server
	consumer  *consumer.Consumer
	processor *processor.Processor
}

func NewApp(ctx context.Context, cfg *factory.Config) (*DnfApp, error) {
	dnf := &DnfApp{
		cfg: cfg,
		wg:  sync.WaitGroup{},
	}
	dnf.SetLogEnable(cfg.GetLogEnable())
	dnf.SetLogLevel(cfg.GetLogLevel())
	dnf.SetReportCaller(cfg.GetLogReportCaller())
	dnf_context.Init()

	processor, err_p := processor.NewProcessor(dnf)
	if err_p != nil {
		return dnf, err_p
	}
	dnf.processor = processor

	consumer, err := consumer.NewConsumer(dnf)
	if err != nil {
		return dnf, err
	}
	dnf.consumer = consumer

	dnf.ctx, dnf.cancel = context.WithCancel(ctx)
	dnf.dnfCtx = dnf_context.GetSelf()

	if dnf.sbiServer, err = sbi.NewServer(dnf); err != nil {
		return nil, err
	}

	DNF = dnf

	return dnf, nil
}

func (a *DnfApp) Start() {
	logger.InitLog.Infoln("Server started")
}

func (a *DnfApp) CancelContext() context.Context {
	return a.ctx
}

func (a *DnfApp) Consumer() *consumer.Consumer {
	return a.consumer
}

func (a *DnfApp) Processor() *processor.Processor {
	return a.processor
}

func (a *DnfApp) Context() *dnf_context.DnfContext {
	return a.dnfCtx
}

func (a *DnfApp) Config() *factory.Config {
	return a.cfg
}

func (a *DnfApp) SetLogEnable(enable bool) {
	logger.MainLog.Infof("Log enable is set to [%v]", enable)
	if enable && logger.Log.Out == os.Stderr {
		return
	} else if !enable && logger.Log.Out == io.Discard {
		return
	}

	a.Config().SetLogEnable(enable)
	if enable {
		logger.Log.SetOutput(os.Stderr)
	} else {
		logger.Log.SetOutput(io.Discard)
	}
}

func (a *DnfApp) SetLogLevel(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		logger.MainLog.Warnf("Log level [%s] is invalid", level)
		return
	}

	logger.MainLog.Infof("Log level is set to [%s]", level)
	if lvl == logger.Log.GetLevel() {
		return
	}

	a.Config().SetLogLevel(level)
	logger.Log.SetLevel(lvl)
}

func (a *DnfApp) SetReportCaller(reportCaller bool) {
	logger.MainLog.Infof("Report Caller is set to [%v]", reportCaller)
	if reportCaller == logger.Log.ReportCaller {
		return
	}

	a.Config().SetLogReportCaller(reportCaller)
	logger.Log.SetReportCaller(reportCaller)
}
