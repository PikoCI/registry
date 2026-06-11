package download

import (
	"context"
	"time"
)

//go:generate go tool mockgen -destination=../mock/download_repository.go -mock_names=Repository=DownloadRepository -package mock github.com/pikoci/registry/pkreg/download Repository

type Repository interface {
	RecordDownload(ctx context.Context, versionID, tokenHash, ipHash string, date time.Time) (bool, error)
}
