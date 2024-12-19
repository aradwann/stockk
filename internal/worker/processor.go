package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"stockk/internal/config"
	"stockk/internal/mail"
	"stockk/internal/repository"

	"github.com/hibiken/asynq"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendAlertEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server         *asynq.Server
	ingredientRepo *repository.IngredientRepository
	mailer         mail.EmailSender
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, ingredientRepo *repository.IngredientRepository, mailer mail.EmailSender) TaskProcessor {
	server := asynq.NewServer(redisOpt, asynq.Config{
		Queues: map[string]int{
			QueueCritical: 6,
			QueueDefault:  3,
			QueueLow:      1,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			slog.LogAttrs(ctx,
				slog.LevelError,
				"process task failed",
				slog.String("err", err.Error()),
				slog.String("type", task.Type()),
				slog.String("payload", string(task.Payload())))
		}),
		Logger: NewLogger(),
	})

	return &RedisTaskProcessor{
		server:         server,
		ingredientRepo: ingredientRepo,
		mailer:         mailer,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(repository.TaskSendAlertEmail, processor.ProcessTaskSendAlertEmail)
	return processor.server.Start(mux)
}

// RunTaskProcessor runs the task processor.
func RunTaskProcessor(config config.Config, redisOpts asynq.RedisClientOpt, ingredientRepo *repository.IngredientRepository) {
	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	taskProcessor := NewRedisTaskProcessor(redisOpts, ingredientRepo, mailer)
	slog.Info("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("%s: %v", "err", err))
		os.Exit(1)
	}
}
