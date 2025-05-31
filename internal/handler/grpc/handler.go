package grpc

import (
	"context"

	"github.com/hoangdv99/morgana/internal/generated/grpc/morgana"
	"github.com/hoangdv99/morgana/internal/logic"
)

type Handler struct {
	morgana.UnimplementedMorganaServiceServer
	accountLogic      logic.Account
	downloadTaskLogic logic.DownloadTask
}

func NewHandler(accountLogic logic.Account, downloadTaskLogic logic.DownloadTask) morgana.MorganaServiceServer {
	return &Handler{
		accountLogic:      accountLogic,
		downloadTaskLogic: downloadTaskLogic,
	}
}

func (a Handler) CreateAccount(ctx context.Context, request *morgana.CreateAccountRequest) (*morgana.CreateAccountResponse, error) {
	// output, err := a.accountLogic.CreateAccount(ctx, logic.CreateAccountParams{
	// 	AccountName: request.Get(),
	// 	Password:    request.Password,
	// })
	panic("unimplemented")
}
func (a Handler) CreateDownloadTask(ctx context.Context, request *morgana.CreateDownloadTaskRequest) (*morgana.CreateDownloadTaskResponse, error) {
	output, err := a.downloadTaskLogic.CreateDownloadTask(ctx, logic.CreateDownloadTaskParams{
		Token:        request.GetToken(),
		DownloadType: request.GetDownloadType(),
		URL:          request.GetUrl(),
	})
	if err != nil {
		return nil, err
	}

	return &morgana.CreateDownloadTaskResponse{
		DownloadTask: &output.DownloadTask,
	}, nil
}

func (a Handler) CreateSession(ctx context.Context, request *morgana.CreateSessionRequest) (*morgana.CreateSessionResponse, error) {
	token, err := a.accountLogic.CreateSession(ctx, logic.CreateSessionParams{
		AccountName: request.GetAccountName(),
		Password:    request.GetPassword(),
	})
	if err != nil {
		return nil, err
	}
	return &morgana.CreateSessionResponse{
		Token: token,
	}, nil
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
// func (a *Handler) mustEmbedUnimplementedGoLoadServiceServer() {
// 	panic("unimplemented")
// }
