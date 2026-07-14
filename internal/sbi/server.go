package sbi

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/free5gc/dnf/internal/logger"
	"github.com/free5gc/dnf/internal/sbi/consumer"
	"github.com/free5gc/dnf/internal/sbi/processor"
	"github.com/free5gc/dnf/pkg/app"
	"github.com/free5gc/dnf/pkg/factory"
	"github.com/free5gc/util/httpwrapper"
	logger_util "github.com/free5gc/util/logger"
)

type ServiceName string

const (
	ServiceName_NDNF_DUMMY ServiceName = "ndnf-dummy"
)

type ServerDnf interface {
	app.App

	Consumer() *consumer.Consumer
	Processor() *processor.Processor
}

type Server struct {
	ServerDnf

	httpServer *http.Server
	router     *gin.Engine
}

func NewServer(dnf ServerDnf) (*Server, error) {
	s := &Server{
		ServerDnf: dnf,
	}

	s.router = newRouter(s)

	cfg := s.Config()
	bindAddr := cfg.GetSbiBindingAddr()
	logger.SBILog.Infof("Binding addr: [%s]", bindAddr)
	var err error
	if s.httpServer, err = httpwrapper.NewHttp2Server(bindAddr, "", s.router); err != nil {
		logger.InitLog.Errorf("Initialize HTTP server failed: %v", err)
		return nil, err
	}
	s.httpServer.ErrorLog = log.New(logger.SBILog.WriterLevel(logrus.ErrorLevel), "HTTP2: ", 0)

	return s, nil
}

func newRouter(s *Server) *gin.Engine {
	router := logger_util.NewGinWithLogrus(logger.GinLog)

	for _, serviceName := range factory.DnfConfig.Configuration.ServiceNameList {
		switch ServiceName(serviceName) {
		case ServiceName_NDNF_DUMMY:
			dnfDummyGroup := router.Group(factory.DnfDummyUriPrefix)
			dnfDummyRoutes := s.getDummyRoutes()
			applyRoutes(dnfDummyGroup, dnfDummyRoutes)

		default:
			logger.SBILog.Warnf("Unsupported service name: %s", serviceName)
		}
	}

	return router
}

func (s *Server) Run(traceCtx context.Context, wg *sync.WaitGroup) error {

	return nil
}
