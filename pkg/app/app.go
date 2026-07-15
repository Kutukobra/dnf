package app

import (
	dnf_context "github.com/free5gc/dnf/internal/context"
	"github.com/free5gc/dnf/pkg/factory"
)

type App interface {
	SetLogEnable(enable bool)
	SetLogLevel(level string)
	SetReportCaller(reportCaller bool)

	Start()
	Terminate()

	Context() *dnf_context.DNFContext
	Config() *factory.Config
}
