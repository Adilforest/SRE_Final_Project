package nats

import (
    "encoding/json"
    "log"

    "github.com/nats-io/nats.go"
)

type OrderProcessedEvent struct {
    OrderID string `json:"order_id"`
    Status  string `json:"status"`
    Message string `json:"message"`
}

func SubscribeOrderProcessed(nc *nats.Conn, handle func(OrderProcessedEvent)) error {
    _, err := nc.Subscribe("order.processed", func(m *nats.Msg) {
        var event OrderProcessedEvent
        if err := json.Unmarshal(m.Data, &event); err != nil {
            log.Printf("Failed to unmarshal order.processed event: %v", err)
            return
        }
        handle(event)
    })
    return err
}