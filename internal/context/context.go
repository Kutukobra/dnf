package context

import (
	"fmt"
	"os"
	"strconv"

	"github.com/free5gc/dnf/internal/logger"
	"github.com/free5gc/dnf/pkg/factory"
	"github.com/free5gc/openapi/models"
)

const (
	ServiceName_NDNF_DUMMY models.ServiceName = "ndnf-dummy"
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
	// Defaults
	context.NrfUri = configuration.NrfUri
	context.NrfCertPem = configuration.NrfCertPem
	context.UriScheme = models.UriScheme(configuration.Sbi.Scheme)
	context.RegisterIPv4 = factory.DnfSbiDefaultIPv4
	context.SBIPort = factory.DnfSbiDefaultPort
	if sbi != nil {
		if sbi.RegisterIPv4 != "" {
			context.RegisterIPv4 = sbi.RegisterIPv4
		}
		if sbi.Port != 0 {
			context.SBIPort = sbi.Port
		}

		if sbi.Scheme == "https" {
			context.UriScheme = models.UriScheme_HTTPS
		} else {
			context.UriScheme = models.UriScheme_HTTP
		}

		context.BindingIPv4 = os.Getenv(sbi.BindingIPv4)
		if context.BindingIPv4 != "" {
			logger.InitLog.Info("Parsing ServerIPv4 address from ENV Variable.")
		} else {
			context.BindingIPv4 = sbi.BindingIPv4
			if context.BindingIPv4 == "" {
				logger.InitLog.Warn("Error parsing ServerIPv4 address as string. Using the 0.0.0.0 address as default.")
				context.BindingIPv4 = "0.0.0.0"
			}
		}
	}

	context.Url = string(context.UriScheme) + "://" + context.RegisterIPv4 + ":" + strconv.Itoa(context.SBIPort)

	context.NfService = make(map[models.ServiceName]models.NrfNfManagementNfService)
	AddNfServices(&context.NfService, config, context)

	fmt.Println("DNF Context: ", context)
}

func AddNfServices(
	serviceMap *map[models.ServiceName]models.NrfNfManagementNfService, config *factory.Config, context *DNFContext,
) {
	var nfService models.NrfNfManagementNfService
	var ipEndPoints []models.IpEndPoint
	var nfServiceVersions []models.NfServiceVersion
	services := *serviceMap

	nfService.ServiceInstanceId = context.NfId
	nfService.ServiceName = models.ServiceName(ServiceName_NDNF_DUMMY)

	var ipEndPoint models.IpEndPoint
	ipEndPoint.Ipv4Address = context.RegisterIPv4
	ipEndPoint.Port = int32(context.SBIPort)
	ipEndPoints = append(ipEndPoints, ipEndPoint)

	var nfServiceVersion models.NfServiceVersion
	nfServiceVersion.ApiFullVersion = config.Info.Version
	nfServiceVersion.ApiVersionInUri = "v1"
	nfServiceVersions = append(nfServiceVersions, nfServiceVersion)

	nfService.Scheme = context.UriScheme
	nfService.NfServiceStatus = models.NfServiceStatus_REGISTERED

	nfService.IpEndPoints = ipEndPoints
	nfService.Versions = nfServiceVersions
	services[ServiceName_NDNF_DUMMY] = nfService
}

func GetSelf() *DNFContext {
	return &dnfContext
}
