package processor

import (
	"github.com/free5gc/dnf/internal/sbi/consumer"
	"github.com/free5gc/dnf/pkg/app"
)

type ProcessorDnf interface {
	app.App

	Consumer() *consumer.Consumer
}

type Processor struct {
	ProcessorDnf
}

func NewProcessor(dnf ProcessorDnf) (*Processor, error) {
	p := &Processor{
		ProcessorDnf: dnf,
	}
	return p, nil
}
