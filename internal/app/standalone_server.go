package app

import (
	"context"
	"syscall"

	"github.com/go-co-op/gocron/v2"
	"github.com/hoangdv99/morgana/internal/configs"
	"github.com/hoangdv99/morgana/internal/handler/consumers"
	"github.com/hoangdv99/morgana/internal/handler/grpc"
	"github.com/hoangdv99/morgana/internal/handler/http"
	"github.com/hoangdv99/morgana/internal/handler/jobs"
	"github.com/hoangdv99/morgana/internal/utils"
	"go.uber.org/zap"
)

type StandaloneServer struct {
	grpcServer                                               grpc.Server
	httpServer                                               http.Server
	rootConsumer                                             consumers.Root
	executeAllPendingDownloadTaskJob                         jobs.ExecuteAllPendingDownloadTask
	updateDownloadingAndFailedDownloadTaskStatusToPendingJob jobs.UpdateDownloadingAndFailedDownloadTaskStatusToPending
	logger                                                   *zap.Logger
	cronConfig                                               configs.Cron
}

func NewStandaloneServer(
	grpcServer grpc.Server,
	httpServer http.Server,
	rootConsumer consumers.Root,
	executeAllPendingDownloadTaskJob jobs.ExecuteAllPendingDownloadTask,
	updateDownloadingAndFailedDownloadTaskStatusToPendingJob jobs.UpdateDownloadingAndFailedDownloadTaskStatusToPending,
	logger *zap.Logger,
	cronConfig configs.Cron,
) *StandaloneServer {
	return &StandaloneServer{
		grpcServer:                       grpcServer,
		httpServer:                       httpServer,
		rootConsumer:                     rootConsumer,
		executeAllPendingDownloadTaskJob: executeAllPendingDownloadTaskJob,
		updateDownloadingAndFailedDownloadTaskStatusToPendingJob: updateDownloadingAndFailedDownloadTaskStatusToPendingJob,
		logger:     logger,
		cronConfig: cronConfig,
	}
}

func (s StandaloneServer) scheduleCronJobs(scheduler gocron.Scheduler) error {
	_, err := scheduler.NewJob(
		gocron.CronJob(s.cronConfig.ExecuteAllPendingDownloadTask.Schedule, true),
		gocron.NewTask(func() {
			err := s.executeAllPendingDownloadTaskJob.Run(context.Background())
			if err != nil {
				s.logger.With(zap.Error(err)).Error("failed to run execute all pending download task job")
			}
		}),
	)
	if err != nil {
		s.logger.With(zap.Error(err)).Error("failed to schedule execute all pending download task job")
		return err
	}
	return nil
}

func (s StandaloneServer) Start() error {
	go func() {
		err := s.grpcServer.Start(context.Background())
		if err != nil {
			s.logger.With(zap.Error(err)).Info("failed to start gRPC server")
		}
	}()

	go func() {
		err := s.httpServer.Start(context.Background())
		if err != nil {
			s.logger.With(zap.Error(err)).Info("failed to start HTTP server")
		}
	}()

	go func() {
		err := s.rootConsumer.Start(context.Background())
		s.logger.With(zap.Error(err)).Info("message queue consumer stopped")
	}()

	utils.BlockUntilSignal(syscall.SIGINT, syscall.SIGTERM)

	return nil
}
