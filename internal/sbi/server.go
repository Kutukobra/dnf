package sbi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/dnf/internal/sbi/consumer"
	"github.com/free5gc/dnf/internal/sbi/processor"
	"github.com/free5gc/dnf/pkg/app"
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

	return s, nil
}
