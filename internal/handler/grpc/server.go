package grpc

import (
	"context"
	"net"

	"github.com/hoangdv99/morgana/internal/generated/github.com/hoangdv99/morgana/morgana"
	"google.golang.org/grpc"
)

type Server struct {
	Start func(ctx context.Context) error
}

type server struct {
	handler morgana.MorganaServiceServer
}

func (s *server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		return err
	}

	defer listener.Close()

	server := grpc.NewServer()
	morgana.RegisterMorganaServiceServer(server, s.handler)
	return server.Serve(listener)
}
