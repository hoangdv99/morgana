package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hoangdv99/morgana/internal/dataaccess/mq/consumer"
	"github.com/hoangdv99/morgana/internal/dataaccess/mq/producer"
	"github.com/hoangdv99/morgana/internal/utils"
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
	logger := utils.LoggerWithContext(ctx, r.logger)

	err := r.mqConsumer.RegisterHandler(
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
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to register handler for DownloadTaskCreated")
		return fmt.Errorf("failed to register handler for DownloadTaskCreated: %w", err)
	}

	return r.mqConsumer.Start(ctx)
}
