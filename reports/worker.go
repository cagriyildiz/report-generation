package reports

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"log/slog"
	"report-generation/config"
)

type Worker struct {
	cfg         config.Config
	builder     *ReportBuilder
	logger      *slog.Logger
	sqsClient   *sqs.Client
	channel     chan types.Message
	concurrency int
}

func NewWorker(cfg config.Config, builder *ReportBuilder, logger *slog.Logger, sqsClient *sqs.Client, maxConcurrency int) *Worker {
	return &Worker{
		cfg:         cfg,
		builder:     builder,
		logger:      logger,
		sqsClient:   sqsClient,
		channel:     make(chan types.Message, maxConcurrency),
		concurrency: maxConcurrency,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	queueUrlOutput, err := w.sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(w.cfg.AWSSQSQueue),
	})
	if err != nil {
		return fmt.Errorf("failed to get url for queue: %s: %w", w.cfg.AWSSQSQueue, err)
	}

	for i := 0; i < w.concurrency; i++ {
		go func(id int) {
			for {
				select {
				case <-ctx.Done():
					w.logger.Error("worker stopped", "goroutine_id", id, "error", ctx.Err())
					return
				case message := <-w.channel:
					if err := w.processMessage(ctx, message, *queueUrlOutput.QueueUrl); err != nil {
						w.logger.Error("failed to process message", "goroutine_id", id, "error", err)
					}
				}
			}
		}(i)
	}

	return nil
}

func (w *Worker) processMessage(ctx context.Context, message types.Message, queueUrl string) error {
	return nil
}
