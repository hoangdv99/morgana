package app

import (
	"context"
	"syscall"

	"github.com/hoangdv99/morgana/internal/handler/grpc"
	"github.com/hoangdv99/morgana/internal/handler/http"
	"github.com/hoangdv99/morgana/internal/utils"
	"go.uber.org/zap"
)

type Server struct {
	grpcServer grpc.Server
	httpServer http.Server
	logger     *zap.Logger
}

func NewServer(grpcServer grpc.Server, httpServer http.Server, logger *zap.Logger) *Server {
	return &Server{
		grpcServer: grpcServer,
		httpServer: httpServer,
		logger:     logger,
	}
}

func (s Server) Start() {
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
}
