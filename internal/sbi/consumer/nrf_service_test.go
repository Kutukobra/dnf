package consumer

import (
	"testing"

	dnf_context "github.com/free5gc/dnf/internal/context"
	"github.com/free5gc/dnf/pkg/app"
	"github.com/free5gc/dnf/pkg/factory"
)

type stubConsumerDnf struct{}

func (s *stubConsumerDnf) SetLogEnable(bool)                {}
func (s *stubConsumerDnf) SetLogLevel(string)               {}
func (s *stubConsumerDnf) SetReportCaller(bool)             {}
func (s *stubConsumerDnf) Start()                           {}
func (s *stubConsumerDnf) Terminate()                       {}
func (s *stubConsumerDnf) Context() *dnf_context.DNFContext { return &dnf_context.DNFContext{} }
func (s *stubConsumerDnf) Config() *factory.Config          { return &factory.Config{} }

var _ app.App = (*stubConsumerDnf)(nil)

func TestGetNFDiscClientInitializesCache(t *testing.T) {
	c, err := NewConsumer(&stubConsumerDnf{})
	if err != nil {
		t.Fatalf("NewConsumer() error = %v", err)
	}

	client := c.nnrfService.getNFDiscClient("http://127.0.0.1:8000")
	if client == nil {
		t.Fatal("expected NF discovery client to be initialized")
	}
}

func TestGetNSSAIReturnsErrorWhenUDMUriMissing(t *testing.T) {
	c, err := NewConsumer(&stubConsumerDnf{})
	if err != nil {
		t.Fatalf("NewConsumer() error = %v", err)
	}

	_, err = c.nudmService.GetNSSAI("imsi-00101", "001", "01")
	if err == nil {
		t.Fatal("expected GetNSSAI to return an error when UDM URI is not configured")
	}
}
