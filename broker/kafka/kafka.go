package kafka

import (
	"fmt"
	"github.com/pkritiotis/outbox"
)

type Broker struct {
}

func (k Broker) Send(event outbox.Message) error {

	fmt.Printf("i'm here %v,%v\n", event, string(event.Body))
	return nil
}