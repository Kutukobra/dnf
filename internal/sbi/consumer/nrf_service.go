package consumer

import (
	"context"
	"sync"

	dnf_context "github.com/free5gc/dnf/internal/context"
	"github.com/free5gc/dnf/internal/logger"
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
	resourceNfUri string, retrieveNfInstanceID string, err error,
) {
	dnfContext := s.consumer.Context()
	client := s.getNFManagementClient(dnfContext.NrfUri)
	nfProfile, err := s.buildNfProfile(dnfContext)
	if err != nil {
		return "", "", errors.Wrap(err, "RegisterNFInstance buildNfProfile()")
	}
}

func (s *nnrfService) buildNfProfile(dnfContext *dnf_context.DNFContext) (
	models.NrfNfManagementNfProfile, error,
) {
	var profile models.NrfNfManagementNfProfile
	profile.NfInstanceId = dnfContext.NfId
	profile.NfType = models.NrfNfManagementNfType("DNF")
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
