package database

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/hoangdv99/morgana/internal/generated/grpc/morgana"
	"go.uber.org/zap"
)

type DownloadTask struct {
	ID             uint64                 `db:"id"`
	AccountID      uint64                 `db:"account_id"`
	DownloadType   morgana.DownloadType   `db:"download_type"`
	URL            string                 `db:"url"`
	DownloadStatus morgana.DownloadStatus `db:"download_status"`
	Metadata       string                 `db:"metadata"`
}

type DownloadTaskDataAccessor interface {
	CreateDownloadTask(ctx context.Context, task DownloadTask) (uint64, error)
	GetDownloadTaskListOfUser(ctx context.Context, userID, offset, limit uint64) ([]DownloadTask, error)
	GetDownloadTaskCountOfUser(ctx context.Context, userID uint64) (uint64, error)
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
	return 1, nil
}

func (d downloadTaskDataAccessor) DeleteDownloadTask(ctx context.Context, id uint64) error {
	panic("unimplemented")
}

func (d downloadTaskDataAccessor) GetDownloadTaskCountOfUser(ctx context.Context, userID uint64) (uint64, error) {
	panic("unimplemented")
}

func (d downloadTaskDataAccessor) GetDownloadTaskListOfUser(ctx context.Context, userID uint64, offset uint64, limit uint64) ([]DownloadTask, error) {
	panic("unimplemented")
}

func (d downloadTaskDataAccessor) UpdateDownloadTask(ctx context.Context, task DownloadTask) error {
	panic("unimplemented")
}

func (d downloadTaskDataAccessor) WithDatabase(database Database) DownloadTaskDataAccessor {
	panic("unimplemented")
}
