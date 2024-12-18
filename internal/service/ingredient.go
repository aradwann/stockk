package service

import (
	"context"

	"stockk/internal/models"
	"stockk/internal/repository"
)

type IngredientService struct {
	ingredientRepo *repository.IngredientRepository
}

func NewIngredientService(ingredientRepo *repository.IngredientRepository) *IngredientService {
	return &IngredientService{ingredientRepo: ingredientRepo}
}

func (is *IngredientService) UpdateIngredientStock(ctx context.Context, ingredients []models.Ingredient) error {
	for _, ingredient := range ingredients {
		// Update stock in database
		if err := is.ingredientRepo.UpdateStock(ctx, nil, ingredient.ID, ingredient.CurrentStock); err != nil {
			return err
		}
	}
	return nil
}

func (is *IngredientService) CheckIngredientLevelsAndAlert(ctx context.Context) error {
	ingredients, err := is.ingredientRepo.GetAllIngredients(ctx)
	if err != nil {
		return err
	}

	for _, ingredient := range ingredients {
		// Calculate percentage
		percentRemaining := (ingredient.CurrentStock / ingredient.TotalStock) * 100

		// Check if below 50% and alert not sent
		if percentRemaining < 50 && !ingredient.AlertSent {
			// Send email alert
			// TODO
			// emailSubject := fmt.Sprintf("Low Stock Alert: %s", ingredient.Name)
			// emailBody := fmt.Sprintf("%s is low on stock. Current stock: %.2f",
			// 	ingredient.Name, ingredient.CurrentStock)

			// if err := is.emailService.SendEmail(emailSubject, emailBody); err != nil {
			// 	log.Printf("Failed to send email for %s: %v", ingredient.Name, err)
			// 	continue
			// }

			// Mark alert as sent in database
			// if err := is.ingredientRepo.MarkAlertSent(ctx, ingredient.ID); err != nil {
			// 	log.Printf("Failed to mark alert sent for %s: %v", ingredient.Name, err)
			// }
		}
	}

	return nil
}
