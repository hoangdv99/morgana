package database

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/hoangdv99/morgana/internal/generated/grpc/morgana"
	"github.com/hoangdv99/morgana/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	TabNameDownloadTasks = goqu.T("download_tasks")
)

const (
	ColNameDownloadTaskID             = "id"
	ColNameDownloadTaskAccountID      = "account_id"
	ColNameDownloadTaskDownloadType   = "download_type"
	ColNameDownloadTaskURL            = "url"
	ColNameDownloadTaskDownloadStatus = "download_status"
	ColNameDownloadTaskMetadata       = "metadata"
)

type DownloadTask struct {
	ID             uint64                 `db:"id" goqu:"skipinsert,skipupdate"`
	AccountID      uint64                 `db:"account_id" goqu:"skipinsert,skipupdate"`
	DownloadType   morgana.DownloadType   `db:"download_type"`
	URL            string                 `db:"url"`
	DownloadStatus morgana.DownloadStatus `db:"download_status"`
	Metadata       JSON                   `db:"metadata"`
}

type DownloadTaskDataAccessor interface {
	CreateDownloadTask(ctx context.Context, task DownloadTask) (uint64, error)
	GetDownloadTaskListOfAccount(ctx context.Context, accountID, offset, limit uint64) ([]DownloadTask, error)
	GetDownloadTaskCountOfAccount(ctx context.Context, accountID uint64) (uint64, error)
	GetDownloadTask(ctx context.Context, id uint64) (DownloadTask, error)
	GetDownloadTaskWithXLock(ctx context.Context, id uint64) (DownloadTask, error)
	UpdateDownloadTask(ctx context.Context, task DownloadTask) error
	DeleteDownloadTask(ctx context.Context, id uint64) error
	WithDatabase(database Database) DownloadTaskDataAccessor
}

type downloadTaskDataAccessor struct {
	database Database
	logger   *zap.Logger
}

func NewDownloadTaskDataAccessor(database *goqu.Database, logger *zap.Logger) DownloadTaskDataAccessor {
	return &downloadTaskDataAccessor{
		database: database,
		logger:   logger,
	}
}

func (d downloadTaskDataAccessor) CreateDownloadTask(ctx context.Context, task DownloadTask) (uint64, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Any("task", task))

	result, err := d.database.
		Insert(TabNameDownloadTasks).
		Rows(task).
		Executor().
		ExecContext(ctx)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create download task")
		return 0, status.Error(codes.Internal, "failed to create download task")
	}

	lastInsertedID, err := result.LastInsertId()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get last inserted id")
		return 0, status.Error(codes.Internal, "failed to get last inserted id")
	}

	return uint64(lastInsertedID), nil
}

func (d downloadTaskDataAccessor) DeleteDownloadTask(ctx context.Context, id uint64) error {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", id))

	_, err := d.database.
		Delete(TabNameDownloadTasks).
		Where(goqu.Ex{ColNameDownloadTaskID: id}).
		Executor().
		ExecContext(ctx)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to delete download task")
		return status.Error(codes.Internal, "failed to delete download task")
	}

	return nil
}

func (d downloadTaskDataAccessor) GetDownloadTaskListOfAccount(ctx context.Context, accountID uint64, offset uint64, limit uint64) ([]DownloadTask, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).
		With(zap.Uint64("account_id", accountID)).
		With(zap.Uint64("offset", offset)).
		With(zap.Uint64("limit", limit))

	downloadTaskList := make([]DownloadTask, 0)
	err := d.database.
		Select().
		From(TabNameDownloadTasks).
		Where(goqu.Ex{ColNameDownloadTaskAccountID: accountID}).
		Offset(uint(offset)).
		Limit(uint(limit)).
		Executor().
		ScanStructsContext(ctx, &downloadTaskList)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get download task list of account")
		return nil, status.Error(codes.Internal, "failed to get download task list of account")
	}

	return downloadTaskList, nil
}

func (d downloadTaskDataAccessor) UpdateDownloadTask(ctx context.Context, task DownloadTask) error {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Any("task", task))

	_, err := d.database.
		Update(TabNameDownloadTasks).
		Set(task).
		Where(goqu.Ex{ColNameDownloadTaskID: task.ID}).
		Executor().
		ExecContext(ctx)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to update download task")
		return status.Error(codes.Internal, "failed to update download task")
	}

	return nil
}

func (d downloadTaskDataAccessor) GetDownloadTask(ctx context.Context, id uint64) (DownloadTask, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", id))
	downloadTask := DownloadTask{}

	found, err := d.database.
		From(TabNameDownloadTasks).
		Where(goqu.Ex{ColNameDownloadTaskID: id}).
		Select().
		ScanStructContext(ctx, &downloadTask)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get download task by id")
		return DownloadTask{}, status.Error(codes.Internal, "failed to get download task by id")
	}

	if !found {
		logger.Error("download task not found")
		return DownloadTask{}, status.Error(codes.NotFound, "download task not found")
	}

	return downloadTask, nil
}

func (d downloadTaskDataAccessor) GetDownloadTaskCountOfAccount(ctx context.Context, accountID uint64) (uint64, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("account_id", accountID))
	count, err := d.database.
		From(TabNameDownloadTasks).
		Where(goqu.Ex{ColNameDownloadTaskAccountID: accountID}).
		CountContext(ctx)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get download task count of account")
		return 0, status.Error(codes.Internal, "failed to get download task count of account")
	}

	return uint64(count), nil
}

func (d downloadTaskDataAccessor) GetDownloadTaskWithXLock(ctx context.Context, id uint64) (DownloadTask, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", id))

	downloadTask := DownloadTask{}
	found, err := d.database.
		Select().
		From(TabNameDownloadTasks).
		Where(goqu.Ex{ColNameDownloadTaskID: id}).
		ForUpdate(goqu.Wait).
		Executor().
		ScanStructContext(ctx, &downloadTask)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get download task with x lock")
		return DownloadTask{}, status.Error(codes.Internal, "failed to get download task with x lock")
	}

	if !found {
		logger.Error("download task not found")
		return DownloadTask{}, status.Error(codes.NotFound, "download task not found")
	}

	return downloadTask, nil
}

func (d downloadTaskDataAccessor) WithDatabase(database Database) DownloadTaskDataAccessor {
	return &downloadTaskDataAccessor{
		database: database,
		logger:   d.logger,
	}
}
