package consumers

import (
	"context"

	"github.com/hoangdv99/morgana/internal/dataaccess/mq/producer"
	"github.com/hoangdv99/morgana/internal/logic"
	"github.com/hoangdv99/morgana/internal/utils"
	"go.uber.org/zap"
)

type DownloadTaskCreated interface {
	Handle(ctx context.Context, event producer.DownloadTaskCreated) error
}

type downloadTaskCreated struct {
	downloadTaskLogic logic.DownloadTask
	logger            *zap.Logger
}

func NewDownloadTaskCreated(
	downloadTaskLogic logic.DownloadTask,
	logger *zap.Logger,
) DownloadTaskCreated {
	return &downloadTaskCreated{
		downloadTaskLogic: downloadTaskLogic,
		logger:            logger,
	}
}

func (d downloadTaskCreated) Handle(ctx context.Context, event producer.DownloadTaskCreated) error {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Any("event", event))
	logger.Info("download task created event received")

	err := d.downloadTaskLogic.ExecuteDownloadTask(ctx, event.ID)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to handle download task created event")
		return err
	}

	return nil
}
