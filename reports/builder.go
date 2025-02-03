package reports

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"report-generation/config"
	"report-generation/db/store"
)

type ReportBuilder struct {
	cfg          *config.Config
	reportsStore *store.ReportsStore
	lozClient    *LozClient
	s3Client     *s3.Client
	logger       *slog.Logger
}

func NewReportBuilder(
	cfg *config.Config,
	reportsStore *store.ReportsStore,
	lozClient *LozClient,
	s3Client *s3.Client,
	logger *slog.Logger,
) *ReportBuilder {
	return &ReportBuilder{
		cfg:          cfg,
		reportsStore: reportsStore,
		lozClient:    lozClient,
		s3Client:     s3Client,
		logger:       logger,
	}
}

func (b *ReportBuilder) Build(ctx context.Context, userId, reportId uuid.UUID) (report *store.Report, err error) {
	report, err = b.reportsStore.GetReportByPrimaryKey(ctx, userId, reportId)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	if report.StartedAt != nil {
		return report, nil
	}

	startedAt := time.Now()
	report.StartedAt = &startedAt

	defer func() {
		b.commit(ctx, err, report)
	}()

	resp, err := b.lozClient.GetMonsters()
	if err != nil {
		return nil, fmt.Errorf("failed to get monsters from api: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no monsters found")
	}
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	csvWriter := csv.NewWriter(gzipWriter)

	header := []string{"name", "id", "category", "description", "image", "common_locations", "drops", "dlc"}
	err = csvWriter.Write(header)
	if err != nil {
		return nil, fmt.Errorf("failed to write csv header: %w", err)
	}

	for _, monster := range resp.Data {
		csvRow := []string{
			monster.Name,
			strconv.Itoa(monster.Id),
			monster.Category,
			monster.Description,
			monster.Image,
			strings.Join(monster.CommonLocations, ", "),
			strings.Join(monster.Drops, ", "),
			strconv.FormatBool(monster.Dlc),
		}

		if err := csvWriter.Write(csvRow); err != nil {
			return nil, fmt.Errorf("failed to write csv row: %w", err)
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("failed to write csv: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	key := fmt.Sprintf("/users/%s/report/%s.csv", userId, reportId)
	_, err = b.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(b.cfg.AWSS3Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buffer.Bytes()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to put object %s: %w", key, err)
	}

	report.OutputFilePath = &key
	completedAt := time.Now()
	report.CompletedAt = &completedAt

	b.logger.Info("successfully generated report",
		"report_id", report.Id,
		"user_id", report.UserId,
		"path", report.OutputFilePath,
	)

	return report, nil
}

func (b *ReportBuilder) commit(ctx context.Context, err error, report *store.Report) {
	if err != nil {
		failedAt := time.Now()
		errMsg := err.Error()
		report.FailedAt = &failedAt
		report.ErrorMessage = &errMsg
	}

	if _, err := b.reportsStore.UpdateReport(ctx, report); err != nil {
		b.logger.Error("failed to update report", "error", err.Error())
	}
}
