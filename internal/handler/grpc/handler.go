package grpc

import (
	"context"

	"github.com/hoangdv99/morgana/internal/generated/grpc/morgana"
)

type Handler struct {
	morgana.UnimplementedMorganaServiceServer
}

func NewHandler() morgana.MorganaServiceServer {
	return &Handler{}
}

func (a *Handler) CreateAccount(context.Context, *morgana.CreateAccountRequest) (*morgana.CreateAccountResponse, error) {
	panic("unimplemented")
}
func (a *Handler) CreateDownloadTask(context.Context, *morgana.CreateDownloadTaskRequest) (*morgana.CreateDownloadTaskResponse, error) {
	panic("unimplemented")
}

// CreateSession implements morgana.GoLoadServiceServer.
func (a *Handler) CreateSession(context.Context, *morgana.CreateSessionRequest) (*morgana.CreateSessionResponse, error) {
	panic("unimplemented")
}

// DeleteDownloadTask implements morgana.GoLoadServiceServer.
func (a *Handler) DeleteDownloadTask(context.Context, *morgana.DeleteDownloadTaskRequest) (*morgana.DeleteDownloadTaskResponse, error) {
	panic("unimplemented")
}

// GetDownloadTaskFile implements morgana.GoLoadServiceServer.
func (a *Handler) GetDownloadTaskFile(*morgana.GetDownloadTaskFileRequest, morgana.MorganaService_GetDownloadTaskFileServer) error {
	panic("unimplemented")
}

// GetDownloadTaskList implements morgana.GoLoadServiceServer.
func (a *Handler) GetDownloadTaskList(context.Context, *morgana.GetDownloadTaskListRequest) (*morgana.GetDownloadTaskListResponse, error) {
	panic("unimplemented")
}

// UpdateDownloadTask implements morgana.GoLoadServiceServer.
func (a *Handler) UpdateDownloadTask(context.Context, *morgana.UpdateDownloadTaskRequest) (*morgana.UpdateDownloadTaskResponse, error) {
	panic("unimplemented")
}

// mustEmbedUnimplementedGoLoadServiceServer implements morgana.GoLoadServiceServer.
func (a *Handler) mustEmbedUnimplementedGoLoadServiceServer() {
	panic("unimplemented")
}
