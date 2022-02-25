package intefaces

import (
	"context"

	"gateway/domain/models"
)

type SubscribeService interface {
	Subscribe(ctx context.Context, clientName, queueName string, symbols []string, dataCh chan<- models.Data) error
	Unsubscribe(queueName string) error
}
