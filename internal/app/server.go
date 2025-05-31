package app

import (
	"context"
	"syscall"

	"github.com/hoangdv99/morgana/internal/handler/consumers"
	"github.com/hoangdv99/morgana/internal/handler/grpc"
	"github.com/hoangdv99/morgana/internal/handler/http"
	"github.com/hoangdv99/morgana/internal/utils"
	"go.uber.org/zap"
)

type Server struct {
	grpcServer   grpc.Server
	httpServer   http.Server
	rootConsumer consumers.Root
	logger       *zap.Logger
}

func NewServer(grpcServer grpc.Server, httpServer http.Server, rootConsumer consumers.Root, logger *zap.Logger) *Server {
	return &Server{
		grpcServer:   grpcServer,
		httpServer:   httpServer,
		rootConsumer: rootConsumer,
		logger:       logger,
	}
}

func (s Server) Start() error {
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
