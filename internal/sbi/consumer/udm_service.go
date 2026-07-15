package consumer

import (
	"sync"

	Nudm_SDM "github.com/free5gc/openapi/udm/SubscriberDataManagement"
)

type nudmService struct {
	consumer *Consumer

	sdmMu sync.RWMutex

	sdmClients map[string]*Nudm_SDM.APIClient
}
