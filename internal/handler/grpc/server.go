package grpc

import (
	"context"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	"github.com/hoangdv99/morgana/internal/configs"
	morgana "github.com/hoangdv99/morgana/internal/generated/morgana/v1"
	"github.com/hoangdv99/morgana/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server interface {
	Start(ctx context.Context) error
}

type server struct {
	handler    morgana.MorganaServiceServer
	grpcConfig configs.GRPC
	logger     *zap.Logger
}

func NewServer(handler morgana.MorganaServiceServer, grpcConfig configs.GRPC, logger *zap.Logger) Server {
	return &server{
		handler:    handler,
		grpcConfig: grpcConfig,
		logger:     logger,
	}
}

func (s *server) Start(ctx context.Context) error {
	logger := utils.LoggerWithContext(ctx, s.logger)

	listener, err := net.Listen("tcp", s.grpcConfig.Address)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to start gRPC server")
		return err
	}

	defer listener.Close()

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			validator.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			validator.StreamServerInterceptor(),
		),
	)
	morgana.RegisterMorganaServiceServer(server, s.handler)
	logger.Info("gRPC server started", zap.String("address", s.grpcConfig.Address))

	return server.Serve(listener)
}
