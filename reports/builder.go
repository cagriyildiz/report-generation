package reports

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"report-generation/db/store"
)

type ReportBuilder struct {
	reportsStore *store.ReportsStore
	LozClient    *LozClient
	s3Client     *s3.Client
}

func NewReportBuilder(reportsStore *store.ReportsStore, lozClient *LozClient, s3Client *s3.Client) *ReportBuilder {
	return &ReportBuilder{
		reportsStore: reportsStore,
		LozClient:    lozClient,
		s3Client:     s3Client,
	}
}

func (r *ReportBuilder) BuildReports(ctx context.Context) error {
	return nil
}
