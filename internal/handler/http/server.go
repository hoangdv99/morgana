package http

import (
	"context"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hoangdv99/morgana/internal/configs"
	"github.com/hoangdv99/morgana/internal/generated/grpc/morgana"
	handlerGPRC "github.com/hoangdv99/morgana/internal/handler/grpc"
	"github.com/hoangdv99/morgana/internal/handler/http/servemuxoptions"
	"github.com/hoangdv99/morgana/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	//nolint:gosec // This is just to specify the cookie name
	AuthTokenCookieName = "MORGANA_AUTH"
)

type Server interface {
	Start(ctx context.Context) error
}

type server struct {
	grpcConfig configs.GRPC
	httpConfig configs.HTTP
	authConfig configs.Auth
	logger     *zap.Logger
}

func NewServer(
	grpcConfig configs.GRPC,
	httpConfig configs.HTTP,
	authConfig configs.Auth,
	logger *zap.Logger,
) Server {
	return &server{
		grpcConfig: grpcConfig,
		httpConfig: httpConfig,
		authConfig: authConfig,
		logger:     logger,
	}
}

func (s server) getGRPCGatewayHandler(ctx context.Context) (http.Handler, error) {
	tokenExpiresInDuration, err := s.authConfig.Token.GetExpiresInDuration()
	if err != nil {
		return nil, err
	}

	grpcMux := runtime.NewServeMux(
		servemuxoptions.WithAuthCookieToAuthMetadata(AuthTokenCookieName, handlerGPRC.AuthTokenMetadataName),
		servemuxoptions.WithAuthMetadataToAuthCookie(handlerGPRC.AuthTokenMetadataName, AuthTokenCookieName, tokenExpiresInDuration),
		servemuxoptions.WithRemoveGoAuthMetadata(handlerGPRC.AuthTokenMetadataName),
	)
	err = morgana.RegisterMorganaServiceHandlerFromEndpoint(
		ctx,
		grpcMux,
		s.grpcConfig.Address,
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	)
	if err != nil {
		return nil, err
	}

	return grpcMux, nil
}

func (s server) Start(ctx context.Context) error {
	logger := utils.LoggerWithContext(ctx, s.logger)

	grpcGatewayHandler, err := s.getGRPCGatewayHandler(ctx)
	if err != nil {
		return err
	}

	httpServer := http.Server{
		Addr:              s.httpConfig.Address,
		Handler:           grpcGatewayHandler,
		ReadHeaderTimeout: time.Minute,
	}

	logger.With(zap.String("address", s.httpConfig.Address)).Info("starting HTTP server")

	return httpServer.ListenAndServe()
}
