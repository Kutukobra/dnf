package processor

import (
	"net/http"

	dnf_context "github.com/free5gc/dnf/internal/context"
	"github.com/free5gc/dnf/internal/logger"
	"github.com/free5gc/openapi/models"
	"github.com/gin-gonic/gin"
)

func (p *Processor) HandleDummyProcess(c *gin.Context) {
	logger.DummyLog.Infof("DUMMY PROCESSING YEAH!!!!")

	p.DummyProcess(c)
}

func (p *Processor) DummyProcess(c *gin.Context) {
	dnfContext := dnf_context.GetSelf()
	nrfUri := dnfContext.NrfUri

	// Discover Itself (whuh)

	targetNfType := models.NrfNfManagementNfType_AF
	requestNfType := models.NrfNfManagementNfType_AF

	searchResult, err := p.Consumer().SendSearchNFInstances(nrfUri, targetNfType, requestNfType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	// Finds NSSAI from dnfconfig.yaml
	nssai, err := p.Consumer().GetNSSAI(dnfContext.SearchSupi, dnfContext.SearchMCC, dnfContext.SearchMNC)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	// Return as JSON
	c.JSON(http.StatusOK, gin.H{
		"searchResult": searchResult,
		"nssai":        nssai,
	})
}
