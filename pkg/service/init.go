package service

import (
	"context"
	"io"
	"os"
	"runtime/debug"
	"sync"

	"github.com/sirupsen/logrus"

	dnf_context "github.com/free5gc/dnf/internal/context"
	"github.com/free5gc/dnf/internal/logger"
	"github.com/free5gc/dnf/internal/sbi"
	"github.com/free5gc/dnf/internal/sbi/consumer"
	"github.com/free5gc/dnf/internal/sbi/processor"
	"github.com/free5gc/dnf/pkg/app"
	"github.com/free5gc/dnf/pkg/factory"
)

var DNF *DnfApp

var _ app.App = &DnfApp{}

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

func (a *DnfApp) Context() *dnf_context.DnfContext {
	return a.dnfCtx
}

func (a *DnfApp) Config() *factory.Config {
	return a.cfg
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

func (a *DnfApp) Start() {
	logger.InitLog.Infoln("Server started")

	a.wg.Add(1)
	go a.listenShutdownEvent()

	if err := a.sbiServer.Run(context.Background(), &a.wg); err != nil {
		logger.MainLog.Fatalf("Run SBI server failed: %+v", err)
	}

	a.WaitRoutineStopped()
}

func (a *DnfApp) listenShutdownEvent() {
	defer func() {
		if p := recover(); p != nil {
			logger.MainLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
		a.wg.Done()
	}()

	<-a.ctx.Done()
	a.terminateProcedure()
}

func (a *DnfApp) Terminate() {
	a.cancel()
}

func (a *DnfApp) terminateProcedure() {
	logger.MainLog.Infof("Terminating DNF...")
	a.CallServerStop()

	// deregister with NRF
	problemDetails, err := a.Consumer().SendDeregisterNFInstance()
	if problemDetails != nil {
		logger.MainLog.Errorf("Deregister NF instance Failed Problem[%+v]", problemDetails)
	} else if err != nil {
		logger.MainLog.Errorf("Deregister NF instance Error[%+v]", err)
	} else {
		logger.MainLog.Infof("Deregister from NRF successfully")
	}
	logger.MainLog.Infof("CHF SBI Server terminated")
}

func (a *DnfApp) CallServerStop() {

}

func (a *DnfApp) WaitRoutineStopped() {
	a.wg.Wait()
	logger.MainLog.Infof("DNF App is terminated")
}
