package nats

import (
	"encoding/json"

	"BikeStoreGolang/services/product-service/internal/logger"

	"github.com/nats-io/nats.go"
)

type OrderProcessedEvent struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Publisher interface {
	PublishOrderProcessed(event OrderProcessedEvent) error
}

type natsPublisher struct {
	nc  *nats.Conn
	log logger.Logger
}

func NewPublisher(nc *nats.Conn, log logger.Logger) Publisher {
	return &natsPublisher{nc: nc, log: log}
}

func (p *natsPublisher) PublishOrderProcessed(event OrderProcessedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		p.log.Errorf("Failed to marshal OrderProcessedEvent: %v", err)
		return err
	}
	err = p.nc.Publish("order.processed", data)
	if err != nil {
		p.log.Errorf("Failed to publish order.processed event: %v", err)
		return err
	}
	p.log.Infof("Published order.processed event: order_id=%s, status=%s", event.OrderID, event.Status)
	return nil
}
