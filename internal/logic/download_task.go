package logic

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/hoangdv99/morgana/internal/dataaccess/database"
	"github.com/hoangdv99/morgana/internal/dataaccess/file"
	"github.com/hoangdv99/morgana/internal/dataaccess/mq/producer"
	morgana "github.com/hoangdv99/morgana/internal/generated/morgana/v1"
	"github.com/hoangdv99/morgana/internal/utils"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateDownloadTaskParams struct {
	Token        string
	DownloadType morgana.DownloadType
	URL          string
}

type CreateDownloadTaskOutput struct {
	DownloadTask *morgana.DownloadTask
}

type GetDownloadTaskListParams struct {
	Token  string
	Offset uint64
	Limit  uint64
}

type GetDownloadTaskListOutput struct {
	DownloadTaskList       []*morgana.DownloadTask
	TotalDownloadTaskCount uint64
}

type UpdateDownloadTaskParams struct {
	Token          string
	DownloadTaskID uint64
	URL            string
}

type UpdateDownloadTaskOutput struct {
	DownloadTask *morgana.DownloadTask
}

type DeleteDownloadTaskParams struct {
	Token          string
	DownloadTaskID uint64
}

type DownloadTask interface {
	CreateDownloadTask(ctx context.Context, params CreateDownloadTaskParams) (CreateDownloadTaskOutput, error)
	GetDownloadTaskList(ctx context.Context, params GetDownloadTaskListParams) (GetDownloadTaskListOutput, error)
	UpdateDownloadTask(ctx context.Context, params UpdateDownloadTaskParams) (UpdateDownloadTaskOutput, error)
	DeleteDownloadTask(ctx context.Context, params DeleteDownloadTaskParams) error
	ExecuteDownloadTask(ctx context.Context, id uint64) error
}

type downloadTask struct {
	tokenLogic                  Token
	accountDataAccessor         database.AccountDataAccessor
	downloadTaskDataAccessor    database.DownloadTaskDataAccessor
	downloadTaskCreatedProducer producer.DownloadTaskCreatedProducer
	goquDatabase                *goqu.Database
	fileClient                  file.Client
	logger                      *zap.Logger
}

func NewDownloadTask(
	tokenLogic Token,
	accountDataAccessor database.AccountDataAccessor,
	downloadTaskDataAccessor database.DownloadTaskDataAccessor,
	downloadTaskCreatedProducer producer.DownloadTaskCreatedProducer,
	goquDatabase *goqu.Database,
	fileClient file.Client,
	logger *zap.Logger,
) DownloadTask {
	return &downloadTask{
		tokenLogic:                  tokenLogic,
		accountDataAccessor:         accountDataAccessor,
		downloadTaskDataAccessor:    downloadTaskDataAccessor,
		downloadTaskCreatedProducer: downloadTaskCreatedProducer,
		goquDatabase:                goquDatabase,
		fileClient:                  fileClient,
		logger:                      logger,
	}
}

func (d downloadTask) databaseDownloadTaskToProtoDownloadTask(
	downloadTask database.DownloadTask,
	account database.Account,
) *morgana.DownloadTask {
	return &morgana.DownloadTask{
		Id: downloadTask.ID,
		Account: &morgana.Account{
			Id:          account.ID,
			AccountName: account.AccountName,
		},
		DownloadType:   downloadTask.DownloadType,
		Url:            downloadTask.URL,
		DownloadStatus: morgana.DownloadStatus_DOWNLOAD_STATUS_PENDING,
	}
}

func (d downloadTask) CreateDownloadTask(ctx context.Context, params CreateDownloadTaskParams) (CreateDownloadTaskOutput, error) {
	accountID, _, err := d.tokenLogic.GetAccountIDAndExpireTime(ctx, params.Token)
	if err != nil {
		return CreateDownloadTaskOutput{}, err
	}

	account, err := d.accountDataAccessor.GetAccountByID(ctx, accountID)
	if err != nil {
		return CreateDownloadTaskOutput{}, err
	}

	downloadTask := database.DownloadTask{
		AccountID:      accountID,
		DownloadType:   params.DownloadType,
		URL:            params.URL,
		DownloadStatus: morgana.DownloadStatus_DOWNLOAD_STATUS_PENDING,
		Metadata: database.JSON{
			Data: make(map[string]interface{}),
		},
	}

	txErr := d.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		downloadTaskID, createDownloadTaskErr := d.downloadTaskDataAccessor.WithDatabase(td).CreateDownloadTask(ctx, downloadTask)
		if createDownloadTaskErr != nil {
			return createDownloadTaskErr
		}

		downloadTask.ID = downloadTaskID
		produceErr := d.downloadTaskCreatedProducer.Produce(ctx, producer.DownloadTaskCreated{
			ID: downloadTaskID,
		})
		if produceErr != nil {
			return produceErr
		}

		return nil
	})

	if txErr != nil {
		return CreateDownloadTaskOutput{}, txErr
	}

	return CreateDownloadTaskOutput{
		DownloadTask: d.databaseDownloadTaskToProtoDownloadTask(downloadTask, account),
	}, nil
}

func (d downloadTask) DeleteDownloadTask(ctx context.Context, params DeleteDownloadTaskParams) error {
	accountID, _, err := d.tokenLogic.GetAccountIDAndExpireTime(ctx, params.Token)
	if err != nil {
		return err
	}

	return d.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		downloadTask, getDownloadTaskWithXLockErr := d.downloadTaskDataAccessor.WithDatabase(td).GetDownloadTaskWithXLock(ctx, params.DownloadTaskID)
		if getDownloadTaskWithXLockErr != nil {
			return getDownloadTaskWithXLockErr
		}

		if downloadTask.AccountID != accountID {
			return status.Error(codes.PermissionDenied, "you do not have permission to delete this download task")
		}

		return d.downloadTaskDataAccessor.WithDatabase(td).DeleteDownloadTask(ctx, params.DownloadTaskID)
	})
}

func (d downloadTask) GetDownloadTaskList(ctx context.Context, params GetDownloadTaskListParams) (GetDownloadTaskListOutput, error) {
	accountID, _, err := d.tokenLogic.GetAccountIDAndExpireTime(ctx, params.Token)
	if err != nil {
		return GetDownloadTaskListOutput{}, err
	}
	account, err := d.accountDataAccessor.GetAccountByID(ctx, accountID)
	if err != nil {
		return GetDownloadTaskListOutput{}, err
	}

	totalDownloadTaskCount, err := d.downloadTaskDataAccessor.GetDownloadTaskCountOfAccount(ctx, accountID)
	if err != nil {
		return GetDownloadTaskListOutput{}, err
	}

	downloadTaskList, err := d.downloadTaskDataAccessor.GetDownloadTaskListOfAccount(ctx, accountID, params.Offset, params.Limit)
	if err != nil {
		return GetDownloadTaskListOutput{}, err
	}

	return GetDownloadTaskListOutput{
		TotalDownloadTaskCount: totalDownloadTaskCount,
		DownloadTaskList: lo.Map(downloadTaskList, func(item database.DownloadTask, _ int) *morgana.DownloadTask {
			return d.databaseDownloadTaskToProtoDownloadTask(item, account)
		}),
	}, nil
}

func (d downloadTask) UpdateDownloadTask(ctx context.Context, params UpdateDownloadTaskParams) (UpdateDownloadTaskOutput, error) {
	accountID, _, err := d.tokenLogic.GetAccountIDAndExpireTime(ctx, params.Token)
	if err != nil {
		return UpdateDownloadTaskOutput{}, err
	}

	account, err := d.accountDataAccessor.GetAccountByID(ctx, accountID)
	if err != nil {
		return UpdateDownloadTaskOutput{}, err
	}

	output := UpdateDownloadTaskOutput{}
	txErr := d.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		downloadTask, getDownloadTaskWithLockErr := d.downloadTaskDataAccessor.WithDatabase(td).GetDownloadTaskWithXLock(ctx, params.DownloadTaskID)
		if getDownloadTaskWithLockErr != nil {
			return getDownloadTaskWithLockErr
		}

		if downloadTask.AccountID != accountID {
			return status.Error(codes.PermissionDenied, "you do not have permission to update this download task")
		}

		downloadTask.URL = params.URL
		output.DownloadTask = d.databaseDownloadTaskToProtoDownloadTask(downloadTask, account)

		return d.downloadTaskDataAccessor.WithDatabase(td).UpdateDownloadTask(ctx, downloadTask)
	})

	if txErr != nil {
		return UpdateDownloadTaskOutput{}, txErr
	}

	return output, nil
}

func (d downloadTask) updateDownloadTaskStatusFromPendingToDownloading(
	ctx context.Context,
	id uint64,
) (bool, database.DownloadTask, error) {
	var (
		logger       = utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", id))
		updated      = false
		downloadTask database.DownloadTask
		err          error
	)

	txErr := d.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		downloadTask, err = d.downloadTaskDataAccessor.WithDatabase(td).GetDownloadTaskWithXLock(ctx, id)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to get download task")
			return err
		}

		if downloadTask.DownloadStatus != morgana.DownloadStatus_DOWNLOAD_STATUS_PENDING {
			logger.Warn("download task is not in pending status, will not execute")
			updated = false
			return nil
		}

		downloadTask.DownloadStatus = morgana.DownloadStatus_DOWNLOAD_STATUS_DOWNLOADING
		err = d.downloadTaskDataAccessor.WithDatabase(td).UpdateDownloadTask(ctx, downloadTask)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to update download task")
			return err
		}

		updated = true

		return nil
	})

	if txErr != nil {
		return false, database.DownloadTask{}, err
	}

	return updated, downloadTask, nil
}

func (d downloadTask) ExecuteDownloadTask(ctx context.Context, id uint64) error {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", id))

	updated, downloadTask, err := d.updateDownloadTaskStatusFromPendingToDownloading(ctx, id)
	if err != nil {
		return err
	}

	if !updated {
		return nil
	}

	var downloader Downloader
	switch downloadTask.DownloadType {
	case morgana.DownloadType_DOWNLOAD_TYPE_HTTP:
		downloader = NewHTTPDownloader(downloadTask.URL, d.logger)
	default:
		logger.With(zap.Any("download_type", downloadTask.DownloadType)).Error("unsupported download type")
		return nil
	}

	fileName := fmt.Sprintf("download_file_%d", id)
	fileWriteCloser, err := d.fileClient.Write(ctx, fileName)
	if err != nil {
		return err
	}

	defer fileWriteCloser.Close()

	metadata, err := downloader.Download(ctx, fileWriteCloser)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to download")
		return err
	}

	metadata["file-name"] = fileName
	downloadTask.DownloadStatus = morgana.DownloadStatus_DOWNLOAD_STATUS_SUCCESS
	downloadTask.Metadata = database.JSON{
		Data: metadata,
	}
	err = d.downloadTaskDataAccessor.UpdateDownloadTask(ctx, downloadTask)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to update download task status to success")
		return err
	}

	logger.Info("download task executed successfully")

	return nil
}
