package consumer

import (
	"sync"

	dnf_context "github.com/free5gc/dnf/internal/context"
	"github.com/free5gc/dnf/internal/logger"
	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	Nudm_SDM "github.com/free5gc/openapi/udm/SubscriberDataManagement"
)

type nudmService struct {
	consumer *Consumer

	sdmMu sync.RWMutex

	sdmClients map[string]*Nudm_SDM.APIClient
}

func (s *nudmService) getSubscriberDMngmntClients(uri string) *Nudm_SDM.APIClient {
	if uri == "" {
		return nil
	}
	s.sdmMu.RLock()

	client, ok := s.sdmClients[uri]
	if ok {
		return client
	}

	configuration := Nudm_SDM.NewConfiguration()
	configuration.SetBasePath(uri)
	client = Nudm_SDM.NewAPIClient(configuration)

	s.sdmMu.RUnlock()
	s.sdmMu.Lock()
	defer s.sdmMu.Unlock()
	s.sdmClients[uri] = client
	return client
}

func (s *nudmService) GetNSSAI(supi string, mcc string, mnc string) (
	*models.ProblemDetails,
	error,
) {
	dnfContext := s.consumer.Context()
	client := s.getSubscriberDMngmntClients(dnfContext.UdmUri)

	getNSSAIRequest := Nudm_SDM.GetNSSAIRequest{
		Supi: &supi,
		PlmnId: &models.PlmnId{
			Mcc: mcc,
			Mnc: mnc,
		},
	}

	ctx, problemDetails, err := dnf_context.GetSelf().GetTokenCtx(models.ServiceName_NUDM_SDM, models.NrfNfManagementNfType_UDM)
	if err != nil {
		return problemDetails, err
	}
	res, err := client.SliceSelectionSubscriptionDataRetrievalApi.GetNSSAI(ctx, &getNSSAIRequest)

	if err != nil {
		if apiErr, ok := err.(openapi.GenericOpenAPIError); ok {
			if errModel, ok := apiErr.Model().(Nudm_SDM.GetNSSAIError); ok {
				problemDetails = &errModel.ProblemDetails
			} else {
				err = openapi.ReportError("openapi error")
			}
		} else {
			return nil, openapi.ReportError("openapi error")
		}
	}

	logger.ConsumerLog.Infof("NSSAI of %s: %v", supi, res.Nssai.SingleNssais[0])

	return problemDetails, err
}
