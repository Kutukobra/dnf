package consumer

import (
	"github.com/free5gc/dnf/pkg/app"
	Nnrf_NFDiscovery "github.com/free5gc/openapi/nrf/NFDiscovery"
	Nnrf_NFManagement "github.com/free5gc/openapi/nrf/NFManagement"
	Nudm_SDM "github.com/free5gc/openapi/udm/SubscriberDataManagement"
)

type ConsumerDnf interface {
	app.App
}

type Consumer struct {
	ConsumerDnf

	*nnrfService
	*nudmService
}

func NewConsumer(dnf ConsumerDnf) (*Consumer, error) {
	c := &Consumer{
		ConsumerDnf: dnf,
	}

	c.nnrfService = &nnrfService{
		consumer:        c,
		nfMngmntClients: make(map[string]*Nnrf_NFManagement.APIClient),
		nfDiscClients:   make(map[string]*Nnrf_NFDiscovery.APIClient),
	}

	c.nudmService = &nudmService{
		consumer:   c,
		sdmClients: make(map[string]*Nudm_SDM.APIClient),
	}

	return c, nil
}
