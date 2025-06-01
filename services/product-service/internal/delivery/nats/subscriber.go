package nats

import (
	"encoding/json"

	"BikeStoreGolang/services/product-service/internal/logger"

	"github.com/nats-io/nats.go"
)

type OrderCreatedEvent struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
	Items   []struct {
		ProductID string  `json:"product_id"`
		Quantity  int32   `json:"quantity"`
		Price     float64 `json:"price"`
	} `json:"items"`
	Total   float64 `json:"total"`
	Address string  `json:"address"`
	Status  string  `json:"status"`
}

func SubscribeOrderCreated(nc *nats.Conn, log logger.Logger, handle func(OrderCreatedEvent)) error {
	_, err := nc.Subscribe("order.created", func(m *nats.Msg) {
		var event OrderCreatedEvent
		if err := json.Unmarshal(m.Data, &event); err != nil {
			log.Warnf("Failed to unmarshal order.created event: %v", err)
			return
		}
		log.Infof("Received order.created event: order_id=%s, user_id=%s, status=%s", event.OrderID, event.UserID, event.Status)
		handle(event)
	})
	if err != nil {
		log.Errorf("Failed to subscribe to order.created: %v", err)
	} else {
		log.Info("Subscribed to order.created events")
	}
	return err
}
