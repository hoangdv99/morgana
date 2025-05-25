package app

import (
	"context"
	"syscall"

	"github.com/hoangdv99/morgana/internal/dataaccess/database"
	"github.com/hoangdv99/morgana/internal/handler/grpc"
	"github.com/hoangdv99/morgana/internal/handler/http"
	"github.com/hoangdv99/morgana/internal/utils"
	"go.uber.org/zap"
)

type Server struct {
	databaseMigrator database.Migrator
	grpcServer       grpc.Server
	httpServer       http.Server
	logger           *zap.Logger
}

func NewServer(databaseMigrator database.Migrator, grpcServer grpc.Server, httpServer http.Server, logger *zap.Logger) *Server {
	return &Server{
		databaseMigrator: databaseMigrator,
		grpcServer:       grpcServer,
		httpServer:       httpServer,
		logger:           logger,
	}
}

func (s Server) Start() error {
	err := s.databaseMigrator.Up(context.Background())
	if err != nil {
		s.logger.With(zap.Error(err)).Error("failed to run database migrations")
		return err
	}

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

	utils.BlockUntilSignal(syscall.SIGINT, syscall.SIGTERM)

	return nil
}
