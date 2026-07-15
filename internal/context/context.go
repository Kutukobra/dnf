package context

import (
	"github.com/free5gc/dnf/internal/logger"
	"github.com/free5gc/dnf/pkg/factory"
	"github.com/free5gc/openapi/models"
)

type DNFContext struct {
	NfId         string
	SBIPort      int
	RegisterIPv4 string
	BindingIPv4  string
	Url          string
	UriScheme    models.UriScheme
	NrfUri       string
	NrfCertPem   string
	NfService    map[models.ServiceName]models.NrfNfManagementNfService
}

var dnfContext DNFContext

func Init() {
	InitDnfContext(&dnfContext)
}

func InitDnfContext(context *DNFContext) {
	config := factory.DnfConfig
	logger.InitLog.Infof("dnfconfig Info: Version[%s] Description[%s]\n", config.Info.Version, config.Info.Description)

	configuration := config.Configuration
	sbi := configuration.Sbi

	context.NfId = config.GetNfInstanceId()
}

func GetSelf() *DNFContext {
	return &dnfContext
}
