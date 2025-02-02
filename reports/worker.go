package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"log/slog"
	"report-generation/config"
)

type Worker struct {
	cfg           *config.Config
	reportBuilder *ReportBuilder
	logger        *slog.Logger
	sqsClient     *sqs.Client
	sqsMsgChannel chan types.Message
	concurrency   int
}

func NewWorker(cfg *config.Config, builder *ReportBuilder, logger *slog.Logger, sqsClient *sqs.Client, maxConcurrency int) *Worker {
	return &Worker{
		cfg:           cfg,
		reportBuilder: builder,
		logger:        logger,
		sqsClient:     sqsClient,
		sqsMsgChannel: make(chan types.Message, maxConcurrency),
		concurrency:   maxConcurrency,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	queueUrlOutput, err := w.sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(w.cfg.AWSSQSQueue),
	})
	if err != nil {
		return fmt.Errorf("failed to get url for queue: %s: %w", w.cfg.AWSSQSQueue, err)
	}

	w.logger.Info("starting worker", "queue", w.cfg.AWSSQSQueue, "queueUrl", queueUrlOutput.QueueUrl)
	for i := 0; i < w.concurrency; i++ {
		go func(id int) {
			w.logger.Info(fmt.Sprintf("starting goroutine #%d", id))
			for {
				select {
				case <-ctx.Done():
					w.logger.Error("worker stopped", "goroutine_id", id, "error", ctx.Err())
					return
				case message := <-w.sqsMsgChannel:
					if err := w.processMessage(ctx, message); err != nil {
						w.logger.Error("failed to process message", "goroutine_id", id, "error", err)
						continue
					}

					if _, err := w.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
						QueueUrl:      queueUrlOutput.QueueUrl,
						ReceiptHandle: message.ReceiptHandle,
					}); err != nil {
						w.logger.Error("failed to delete message", "goroutine_id", id, "error", err)
					}
				}
			}
		}(i)
	}

	for {
		messageOutput, err := w.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            queueUrlOutput.QueueUrl,
			MaxNumberOfMessages: int32(w.concurrency + 1),
		})
		if err != nil {
			w.logger.Error("failed to receive message", "error", err)
			if ctx.Err() != nil {
				return ctx.Err()
			}
		}

		if len(messageOutput.Messages) == 0 {
			continue
		}

		for _, message := range messageOutput.Messages {
			w.sqsMsgChannel <- message
		}
	}
}

func (w *Worker) processMessage(ctx context.Context, message types.Message) error {
	w.logger.Info("processing message", "message_id", message.MessageId)
	if message.Body == nil || *message.Body == "" {
		return fmt.Errorf("message body is empty")
	}

	var sqsMessage SqsMessage
	if err := json.Unmarshal([]byte(*message.Body), &sqsMessage); err != nil {
		return err
	}

	if _, err := w.reportBuilder.Build(ctx, sqsMessage.UserId, sqsMessage.ReportId); err != nil {
		return err
	}

	return nil
}
