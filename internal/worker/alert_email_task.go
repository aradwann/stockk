package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"stockk/internal/repository"
	"strings"

	"github.com/hibiken/asynq"
)

func (processor *RedisTaskProcessor) ProcessTaskSendAlertEmail(ctx context.Context, task *asynq.Task) error {
	var payload repository.PayloadSendAlertEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", asynq.SkipRetry)
	}

	// Prepare the email content
	contentBuilder := strings.Builder{}
	contentBuilder.WriteString(`Hello,<br/>
	Thank you for being a valued member of the Stockk community!<br/>
	The following ingredients are running low on stock:<br/><ul>`)

	for _, ingredient := range payload.Ingredients {
		// Calculate the remaining percentage of the stock
		percentRemaining := (ingredient.CurrentStock / ingredient.TotalStock) * 100

		// Add warning details to the email content
		contentBuilder.WriteString(fmt.Sprintf(
			`<li>%s: %.2f%% remaining</li>`,
			ingredient.Name, percentRemaining,
		))

		// Mark the alert as sent for the ingredient
		err := processor.ingredientRepo.MarkAlertSent(ctx, ingredient.ID)
		if err != nil {
			return fmt.Errorf("failed to mark alert as sent for ingredient ID %d: %w", ingredient.ID, err)
		}
	}

	contentBuilder.WriteString("</ul><br/>Please replenish these ingredients soon to avoid disruption.<br/>Best regards,<br/>The Stockk Team")

	// Convert the content to a string
	content := contentBuilder.String()

	// Send the email
	testMerchantEmail := "aradwann@proton.me"
	subject := "Stockk Alert: Low Stock Warning"
	to := []string{testMerchantEmail} // Test email for demonstration purposes
	err := processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send warning email: %w", err)
	}

	// Log the processed task
	slog.LogAttrs(ctx,
		slog.LevelInfo,
		"processed task",
		slog.String("type", task.Type()),
		slog.String("payload", string(task.Payload())),
		slog.String("email", testMerchantEmail),
	)

	return nil
}
