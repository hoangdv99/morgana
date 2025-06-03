package jobs

import (
	"context"

	"github.com/hoangdv99/morgana/internal/logic"
)

type UpdateDownloadingAndFailedDownloadTaskStatusToPending interface {
	Run(ctx context.Context) error
}

type updateDownloadingAndFailedDownloadTaskStatusToPendingJob struct {
	downloadTaskLogic logic.DownloadTask
}

func NewUpdateDownloadingAndFailedDownloadTaskStatusToPending(
	downloadTaskLogic logic.DownloadTask,
) UpdateDownloadingAndFailedDownloadTaskStatusToPending {
	return &updateDownloadingAndFailedDownloadTaskStatusToPendingJob{
		downloadTaskLogic: downloadTaskLogic,
	}
}

func (u updateDownloadingAndFailedDownloadTaskStatusToPendingJob) Run(ctx context.Context) error {
	return u.downloadTaskLogic.UpdateDownloadingAndFailedDownloadTaskStatusToPending(ctx)
}
