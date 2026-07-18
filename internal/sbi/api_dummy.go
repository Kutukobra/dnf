package sbi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) getDummyRoutes() []Route {
	return []Route{
		{
			Name:    "Index",
			Method:  http.MethodGet,
			Pattern: "/",
			APIFunc: s.HTTPDummyMessage,
		},
		{
			Name:    "Dummy Process",
			Method:  http.MethodGet,
			Pattern: "/dummy",
			APIFunc: s.HTTPDummyProcess,
		},
	}
}

func (s *Server) HTTPDummyMessage(c *gin.Context) {
	c.String(http.StatusOK, "Hello DNF!")
}

func (s *Server) HTTPDummyProcess(c *gin.Context) {
	s.Processor().HandleDummyProcess(c)
}
