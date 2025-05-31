package consumers

import (
	"context"
	"encoding/json"

	"github.com/hoangdv99/morgana/internal/dataaccess/mq/consumer"
	"github.com/hoangdv99/morgana/internal/dataaccess/mq/producer"
	"go.uber.org/zap"
)

type Root interface {
	Start(ctx context.Context) error
}

type root struct {
	downloadTaskCreatedHandler DownloadTaskCreated
	mqConsumer                 consumer.Consumer
	logger                     *zap.Logger
}

func NewRoot(
	downloadTaskCreatedHandler DownloadTaskCreated,
	mqConsumer consumer.Consumer,
	logger *zap.Logger,
) Root {
	return &root{
		downloadTaskCreatedHandler: downloadTaskCreatedHandler,
		mqConsumer:                 mqConsumer,
		logger:                     logger,
	}
}

func (r root) Start(ctx context.Context) error {
	r.mqConsumer.RegisterHandler(
		producer.MessageQueueDownloadTaskCreated,
		func(ctx context.Context, queueName string, payload []byte) error {
			var event producer.DownloadTaskCreated
			err := json.Unmarshal(payload, &event)
			if err != nil {
				return err
			}
			return r.downloadTaskCreatedHandler.Handle(ctx, event)
		},
	)

	return r.mqConsumer.Start(ctx)
}
