package nats

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type OrderCreatedEvent struct {
	OrderID string      `json:"order_id"`
	UserID  string      `json:"user_id"`
	Items   interface{} `json:"items"`
	Total   float64     `json:"total"`
	Address string      `json:"address"`
	Status  string      `json:"status"`
}

type Publisher struct {
	nc *nats.Conn
}

func NewPublisher(nc *nats.Conn) *Publisher {
	return &Publisher{nc: nc}
}

func (p *Publisher) PublishOrderCreated(event OrderCreatedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.nc.Publish("order.created", data)
}

func (p *Publisher) PublishOrderApproved(event OrderCreatedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.nc.Publish("order.approved", data)
}
