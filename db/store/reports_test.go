package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReportsStore(t *testing.T) {
	testDB := NewTestDB(t)
	cleanup := testDB.Setup(t)
	t.Cleanup(func() {
		cleanup(t)
	})

	now := time.Now()
	ctx := context.Background()

	userStore := NewUserStore(testDB.DB)
	user, err := userStore.CreateUser(ctx, "test@test.com", "test")
	require.NoError(t, err)

	reportsStore := NewReportsStore(testDB.DB)
	report, err := reportsStore.CreateReport(ctx, user.Id, "test")
	require.NoError(t, err)

	require.Equal(t, user.Id, report.UserId)
	require.Equal(t, "test", report.ReportType)
	require.True(t, report.CreatedAt.After(now))

	startedAt := report.CreatedAt.Add(time.Second)
	completedAt := report.CreatedAt.Add(2 * time.Second)
	failedAt := report.CreatedAt.Add(3 * time.Second)
	errMsg := "error"
	downloadUrl := "http://localhost:8080/reports"
	outputPath := "s3://reports-test/reports"
	downloadUrlExpiresAt := report.CreatedAt.Add(4 * time.Second)

	report.ReportType = "ex"
	report.OutputFilePath = &outputPath
	report.DownloadUrl = &downloadUrl
	report.DownloadUrlExpiresAt = &downloadUrlExpiresAt
	report.StartedAt = &startedAt
	report.CompletedAt = &completedAt
	report.FailedAt = &failedAt
	report.ErrorMessage = &errMsg

	updatedRecord, err := reportsStore.UpdateReport(ctx, report)
	require.NoError(t, err)
	require.Equal(t, report, updatedRecord)

	gotReport, err := reportsStore.GetReportByPrimaryKey(ctx, report.UserId, report.Id)
	require.NoError(t, err)
	require.Equal(t, updatedRecord, gotReport)
}
