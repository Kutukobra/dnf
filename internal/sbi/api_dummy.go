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
			APIFunc: func(c *gin.Context) {
				c.String(http.StatusOK, "Hello DNF!")
			},
		},
	}
}
