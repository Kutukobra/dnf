package consumer

import (
	"context"
	"strings"
	"sync"
	"time"

	dnf_context "github.com/free5gc/dnf/internal/context"
	"github.com/free5gc/dnf/internal/logger"
	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	Nnrf_NFDiscovery "github.com/free5gc/openapi/nrf/NFDiscovery"
	Nnrf_NFManagement "github.com/free5gc/openapi/nrf/NFManagement"
	"github.com/pkg/errors"
)

type nnrfService struct {
	consumer *Consumer

	nfMngmntMu sync.RWMutex
	nfDiscMu   sync.RWMutex

	nfMngmntClients map[string]*Nnrf_NFManagement.APIClient
	nfDiscClients   map[string]*Nnrf_NFDiscovery.APIClient
}

func (s *nnrfService) getNFManagementClient(uri string) *Nnrf_NFManagement.APIClient {
	if uri == "" {
		return nil
	}
	s.nfMngmntMu.RLock()
	client, ok := s.nfMngmntClients[uri]
	if ok {
		s.nfMngmntMu.RUnlock()
		return client
	}

	configuration := Nnrf_NFManagement.NewConfiguration()
	configuration.SetBasePath(uri)
	client = Nnrf_NFManagement.NewAPIClient(configuration)

	s.nfMngmntMu.RUnlock()
	s.nfMngmntMu.Lock()
	defer s.nfMngmntMu.Unlock()
	s.nfMngmntClients[uri] = client
	return client
}

func (s *nnrfService) SendDeregisterNFInstance() (*models.ProblemDetails, error) {
	logger.ConsumerLog.Infof("[DNF] Send Deregister NFInstance")
	return nil, nil
}

func (s *nnrfService) RegisterNFInstance(ctx context.Context) (
	string, string, error,
) {
	dnfContext := s.consumer.Context()
	client := s.getNFManagementClient(dnfContext.NrfUri)
	nfProfile, err := s.buildNfProfile(dnfContext)
	if err != nil {
		return "", "", errors.Wrap(err, "RegisterNFInstance buildNfProfile()")
	}

	var nf models.NrfNfManagementNfProfile
	var res *Nnrf_NFManagement.RegisterNFInstanceResponse
	registerNFInstanceRequest := &Nnrf_NFManagement.RegisterNFInstanceRequest{
		NfInstanceID:             &dnfContext.NfId,
		NrfNfManagementNfProfile: &nfProfile,
	}

	var resourceNrfUri string
	var retrieveNfInstanceID string
	for {
		select {
		case <-ctx.Done():
			return "", "", errors.Errorf("contex cancel before RegiserNFInstance")
		default:
		}
		res, err = client.NFInstanceIDDocumentApi.RegisterNFInstance(ctx, registerNFInstanceRequest)
		if err != nil || res == nil {
			logger.ConsumerLog.Errorf("DNF register to NRF Error[%v]", err)
			if errorResponse, ok := err.(openapi.GenericOpenAPIError); ok {
				if apiError, ok := errorResponse.Model().(Nnrf_NFManagement.RegisterNFInstanceError); ok {
					logger.ConsumerLog.Errorf("%v", apiError.ProblemDetails.Detail)
				}
			}
			time.Sleep(2 * time.Second)
			continue
		}
		nf = res.NrfNfManagementNfProfile

		if res.Location == "" {
			break
		} else {
			resourceUri := res.Location
			resourceNrfUri = resourceUri[:strings.Index(resourceUri, "/nnrf-nfm/")]
			retrieveNfInstanceID = resourceUri[strings.LastIndex(resourceUri, "/")+1:]

			oauth2 := false
			if nf.CustomInfo != nil {
				v, ok := nf.CustomInfo["oauth2"].(bool)
				if ok {
					oauth2 = v
					logger.MainLog.Infoln("OAuth2 setting receive from NRF: ", oauth2)
				}
			}
			dnf_context.GetSelf().OAuth2Required = oauth2
			if oauth2 && dnf_context.GetSelf().NrfCertPem == "" {
				logger.CfgLog.Error("OAuth2 enable but no nrfCertPem provided in config.")
			}

			break
		}
	}
	return resourceNrfUri, retrieveNfInstanceID, err
}

func (s *nnrfService) buildNfProfile(dnfContext *dnf_context.DNFContext) (
	models.NrfNfManagementNfProfile, error,
) {
	var profile models.NrfNfManagementNfProfile
	profile.NfInstanceId = dnfContext.NfId
	profile.NfType = models.NrfNfManagementNfType_AF
	profile.NfStatus = models.NrfNfManagementNfStatus_REGISTERED
	profile.Ipv4Addresses = append(profile.Ipv4Addresses, dnfContext.RegisterIPv4)

	services := []models.NrfNfManagementNfService{}
	for _, nfService := range dnfContext.NfService {
		services = append(services, nfService)
	}
	if len(services) > 0 {
		profile.NfServices = services
	}

	return profile, nil
}
