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
		DownloadTask: output.DownloadTask,
	}, nil
}

func (a Handler) CreateSession(ctx context.Context, request *morgana.CreateSessionRequest) (*morgana.CreateSessionResponse, error) {
	output, err := a.accountLogic.CreateSession(ctx, logic.CreateSessionParams{
		AccountName: request.GetAccountName(),
		Password:    request.GetPassword(),
	})
	if err != nil {
		return nil, err
	}
	return &morgana.CreateSessionResponse{
		Account: output.Account,
		Token:   output.Token,
	}, nil
}

func (a Handler) DeleteDownloadTask(ctx context.Context, request *morgana.DeleteDownloadTaskRequest) (*morgana.DeleteDownloadTaskResponse, error) {
	err := a.downloadTaskLogic.DeleteDownloadTask(ctx, logic.DeleteDownloadTaskParams{
		Token:          request.GetToken(),
		DownloadTaskID: request.GetDownloadTaskId(),
	})
	if err != nil {
		return nil, err
	}

	return &morgana.DeleteDownloadTaskResponse{}, nil
}

// GetDownloadTaskFile implements morgana.GoLoadServiceServer.
func (a *Handler) GetDownloadTaskFile(*morgana.GetDownloadTaskFileRequest, morgana.MorganaService_GetDownloadTaskFileServer) error {
	panic("unimplemented")
}

func (a Handler) GetDownloadTaskList(ctx context.Context, request *morgana.GetDownloadTaskListRequest) (*morgana.GetDownloadTaskListResponse, error) {
	output, err := a.downloadTaskLogic.GetDownloadTaskList(ctx, logic.GetDownloadTaskListParams{
		Token:  request.GetToken(),
		Offset: request.GetOffset(),
		Limit:  request.GetLimit(),
	})
	if err != nil {
		return nil, err
	}

	return &morgana.GetDownloadTaskListResponse{
		DownloadTaskList:      output.DownloadTaskList,
		ToalDownloadTaskCount: output.TotalDownloadTaskCount,
	}, nil
}

func (a Handler) UpdateDownloadTask(ctx context.Context, request *morgana.UpdateDownloadTaskRequest) (*morgana.UpdateDownloadTaskResponse, error) {
	output, err := a.downloadTaskLogic.UpdateDownloadTask(ctx, logic.UpdateDownloadTaskParams{
		Token:          request.GetToken(),
		DownloadTaskID: request.GetDownloadTaskId(),
		URL:            request.GetUrl(),
	})
	if err != nil {
		return nil, err
	}

	return &morgana.UpdateDownloadTaskResponse{
		DownloadTask: output.DownloadTask,
	}, nil
}

// mustEmbedUnimplementedGoLoadServiceServer implements morgana.GoLoadServiceServer.
// func (a *Handler) mustEmbedUnimplementedGoLoadServiceServer() {
// 	panic("unimplemented")
// }
