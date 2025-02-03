package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ReportsStore struct {
	db *sqlx.DB
}

func NewReportsStore(db *sql.DB) *ReportsStore {
	return &ReportsStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type Report struct {
	UserId               uuid.UUID  `db:"user_id"`
	Id                   uuid.UUID  `db:"id"`
	ReportType           string     `db:"report_type"`
	OutputFilePath       *string    `db:"output_file_path"`
	DownloadUrl          *string    `db:"download_url"`
	DownloadUrlExpiresAt *time.Time `db:"download_url_expires_at"`
	ErrorMessage         *string    `db:"error_message"`
	CreatedAt            time.Time  `db:"created_at"`
	StartedAt            *time.Time `db:"started_at"`
	FailedAt             *time.Time `db:"failed_at"`
	CompletedAt          *time.Time `db:"completed_at"`
}

func (r *Report) IsDone() bool {
	return r.FailedAt != nil || r.CompletedAt != nil
}

func (r *Report) Status() string {
	switch {
	case r.StartedAt == nil:
		return "requested"
	case r.StartedAt != nil && !r.IsDone():
		return "processing"
	case r.CompletedAt != nil:
		return "completed"
	case r.FailedAt != nil:
		return "failed"
	}
	return "unknown"
}

func (s *ReportsStore) CreateReport(ctx context.Context, userId uuid.UUID, reportType string) (*Report, error) {
	const query = `INSERT INTO reports (user_id, report_type) VALUES ($1, $2) RETURNING *`

	var report Report
	err := s.db.GetContext(ctx, &report, query, userId, reportType)
	if err != nil {
		return nil, fmt.Errorf("failed to insert report: %w", err)
	}

	return &report, nil
}

func (s *ReportsStore) UpdateReport(ctx context.Context, report *Report) (*Report, error) {
	const query = `UPDATE reports 
				   SET report_type = $1,
				       output_file_path = $2, 
				       download_url = $3, 
				       download_url_expires_at = $4, 
				       error_message = $5, 
				       started_at = $6, 
				       failed_at = $7, 
				       completed_at = $8 
				   WHERE user_id = $9 and id = $10 RETURNING *`

	var updatedReport Report
	err := s.db.GetContext(ctx, &updatedReport, query,
		report.ReportType,
		report.OutputFilePath,
		report.DownloadUrl,
		report.DownloadUrlExpiresAt,
		report.ErrorMessage,
		report.StartedAt,
		report.FailedAt,
		report.CompletedAt,
		report.UserId,
		report.Id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update report: %w", err)
	}

	return &updatedReport, nil
}

func (s *ReportsStore) GetReportByPrimaryKey(ctx context.Context, userId, id uuid.UUID) (*Report, error) {
	const query = `SELECT * FROM reports WHERE user_id = $1 AND id = $2`

	var report Report
	err := s.db.GetContext(ctx, &report, query, userId, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	return &report, nil
}
