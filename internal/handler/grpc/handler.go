package grpc

import (
	"context"
	"errors"
	"io"

	"github.com/hoangdv99/morgana/internal/configs"
	morgana "github.com/hoangdv99/morgana/internal/generated/morgana/v1"
	"github.com/hoangdv99/morgana/internal/logic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	//nolint:gosec // This is just to specify the metadata name
	AuthTokenMetadataName = "MORGANA_AUTH"
)

type Handler struct {
	morgana.UnimplementedMorganaServiceServer
	accountLogic                                 logic.Account
	downloadTaskLogic                            logic.DownloadTask
	getDownloadTaskFileResponseBufferSizeInBytes uint64
}

func NewHandler(accountLogic logic.Account, downloadTaskLogic logic.DownloadTask, grpcConfig configs.GRPC) (morgana.MorganaServiceServer, error) {
	getDownloadTaskFileResponseBufferSizeInBytes, err := grpcConfig.GetDownloadTaskFile.GetResponseBufferSizeInBytes()
	if err != nil {
		return nil, err
	}

	return &Handler{
		accountLogic:      accountLogic,
		downloadTaskLogic: downloadTaskLogic,
		getDownloadTaskFileResponseBufferSizeInBytes: getDownloadTaskFileResponseBufferSizeInBytes,
	}, nil
}

func (a Handler) CreateAccount(ctx context.Context, request *morgana.CreateAccountRequest) (*morgana.CreateAccountResponse, error) {
	output, err := a.accountLogic.CreateAccount(ctx, logic.CreateAccountParams{
		AccountName: request.GetAccountName(),
		Password:    request.GetPassword(),
	})
	if err != nil {
		return nil, err
	}

	return &morgana.CreateAccountResponse{
		AccountId: output.ID,
	}, nil
}

func (a Handler) getAuthTokenMetadata(ctx context.Context) string {
	metadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	metadataValues := metadata.Get(AuthTokenMetadataName)
	if len(metadataValues) == 0 {
		return ""
	}

	return metadataValues[0]
}

func (a Handler) CreateDownloadTask(ctx context.Context, request *morgana.CreateDownloadTaskRequest) (*morgana.CreateDownloadTaskResponse, error) {
	output, err := a.downloadTaskLogic.CreateDownloadTask(ctx, logic.CreateDownloadTaskParams{
		Token:        a.getAuthTokenMetadata(ctx),
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

	err = grpc.SetHeader(ctx, metadata.Pairs(AuthTokenMetadataName, output.Token))
	if err != nil {
		return nil, err
	}

	return &morgana.CreateSessionResponse{
		Account: output.Account,
	}, nil
}

func (a Handler) DeleteDownloadTask(ctx context.Context, request *morgana.DeleteDownloadTaskRequest) (*morgana.DeleteDownloadTaskResponse, error) {
	err := a.downloadTaskLogic.DeleteDownloadTask(ctx, logic.DeleteDownloadTaskParams{
		Token:          a.getAuthTokenMetadata(ctx),
		DownloadTaskID: request.GetDownloadTaskId(),
	})
	if err != nil {
		return nil, err
	}

	return &morgana.DeleteDownloadTaskResponse{}, nil
}

func (a Handler) GetDownloadTaskFile(request *morgana.GetDownloadTaskFileRequest, server morgana.MorganaService_GetDownloadTaskFileServer) error {
	outputReader, err := a.downloadTaskLogic.GetDownloadTaskFile(server.Context(), logic.GetDownloadTaskFileParams{
		Token:          a.getAuthTokenMetadata(server.Context()),
		DownloadTaskID: request.GetDownloadTaskId(),
	})
	if err != nil {
		return err
	}

	defer outputReader.Close()

	for {
		dataBuffer := make([]byte, a.getDownloadTaskFileResponseBufferSizeInBytes)
		readByteCount, readErr := outputReader.Read(dataBuffer)
		if readByteCount > 0 {
			sendErr := server.Send(&morgana.GetDownloadTaskFileResponse{
				Data: dataBuffer[:readByteCount],
			})
			if sendErr != nil {
				return sendErr
			}
			continue
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return readErr
		}
	}

	return nil
}

func (a Handler) GetDownloadTaskList(ctx context.Context, request *morgana.GetDownloadTaskListRequest) (*morgana.GetDownloadTaskListResponse, error) {
	output, err := a.downloadTaskLogic.GetDownloadTaskList(ctx, logic.GetDownloadTaskListParams{
		Token:  a.getAuthTokenMetadata(ctx),
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
		Token:          a.getAuthTokenMetadata(ctx),
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
