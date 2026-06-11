package db

import (
	"context"
	"fmt"
	"time"

	"github.com/cycloidio/sqlr"
)

type DownloadRepository struct {
	querier sqlr.Querier
}

func NewDownloadRepository(db sqlr.Querier) *DownloadRepository {
	return &DownloadRepository{querier: db}
}

func (r *DownloadRepository) RecordDownload(ctx context.Context, versionID, tokenHash, ipHash string, date time.Time) (bool, error) {
	_, err := r.querier.ExecContext(ctx, `
		INSERT INTO download_log(version_id, token_hash, ip_hash, download_date)
		VALUES (?, ?, ?, ?)
	`, versionID, tokenHash, ipHash, dateToString(date))
	if err != nil {
		if isUniqueViolation(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to record download: %w", err)
	}
	return true, nil
}
