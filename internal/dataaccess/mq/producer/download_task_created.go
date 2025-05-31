package producer

import (
	"context"
	"encoding/json"

	"github.com/hoangdv99/morgana/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MessageQueueDownloadTaskCreated = "download_task_created"
)

type DownloadTaskCreated struct {
	ID uint64 `json:"id"`
}

type DownloadTaskCreatedProducer interface {
	Produce(ctx context.Context, event DownloadTaskCreated) error
}

type downloadTaskCreatedProducer struct {
	client Client
	logger *zap.Logger
}

func NewDownloadTaskCreatedProducer(
	client Client,
	logger *zap.Logger,
) DownloadTaskCreatedProducer {
	return &downloadTaskCreatedProducer{
		client: client,
		logger: logger,
	}
}

func (d downloadTaskCreatedProducer) Produce(ctx context.Context, event DownloadTaskCreated) error {
	logger := utils.LoggerWithContext(ctx, d.logger)

	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to marshal DownloadTaskCreated event")
		return status.Errorf(codes.Internal, "failed to marshal DownloadTaskCreated event: %v", err)
	}

	err = d.client.Produce(ctx, MessageQueueDownloadTaskCreated, eventBytes)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to produce DownloadTaskCreated event")
		return status.Errorf(codes.Internal, "failed to produce DownloadTaskCreated event: %v", err)
	}

	return nil
}
