package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"stockk/internal/models"

	"github.com/hibiken/asynq"
)

const TaskSendAlertEmail = "task:send_alert_email"

type PayloadSendAlertEmail struct {
	Ingredients []models.Ingredient `json:"ingredients"`
}

type TaskQueueRepository struct {
	client *asynq.Client
}

func NewTaskQueueRepository(client *asynq.Client) *TaskQueueRepository {
	return &TaskQueueRepository{client: client}
}

func (r *TaskQueueRepository) EnqueueAlertEmailTask(ctx context.Context,
	payload *PayloadSendAlertEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}
	task := asynq.NewTask(TaskSendAlertEmail, jsonPayload, opts...)
	info, err := r.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	slog.LogAttrs(context.Background(),
		slog.LevelInfo,
		"enqueued task",
		slog.String("type", task.Type()),
		slog.String("payload", string(task.Payload())),
		slog.String("queue", info.Queue),
		slog.Int("max_retry", info.MaxRetry),
	)
	return nil
}
