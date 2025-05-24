package grpc

import (
	"context"

	"github.com/hoangdv99/morgana/internal/generated/grpc/morgana"
	"github.com/hoangdv99/morgana/internal/logic"
)

type Handler struct {
	morgana.UnimplementedMorganaServiceServer
	accountLogic logic.Account
}

func NewHandler(accountLogic logic.Account) morgana.MorganaServiceServer {
	return &Handler{
		accountLogic: accountLogic,
	}
}

func (a Handler) CreateAccount(ctx context.Context, request *morgana.CreateAccountRequest) (*morgana.CreateAccountResponse, error) {
	// output, err := a.accountLogic.CreateAccount(ctx, logic.CreateAccountParams{
	// 	AccountName: request.Get(),
	// 	Password:    request.Password,
	// })
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
