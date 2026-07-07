package consumer

import "github.com/free5gc/dnf/pkg/app"

type ConsumerDnf interface {
	app.App
}

type Consumer struct {
	ConsumerDnf
}

func NewConsumer(dnf ConsumerDnf) (*Consumer, error) {
	c := &Consumer{
		ConsumerDnf: dnf,
	}

	return c, nil
}
