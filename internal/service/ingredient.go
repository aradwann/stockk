package service

import (
	"context"

	"stockk/internal/models"
	"stockk/internal/repository"
)

type IngredientService struct {
	ingredientRepo *repository.IngredientRepository
	taskRepo       *repository.TaskQueueRepository
}

func NewIngredientService(ingredientRepo *repository.IngredientRepository, taskRepo *repository.TaskQueueRepository) *IngredientService {
	return &IngredientService{ingredientRepo: ingredientRepo, taskRepo: taskRepo}
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
	ingredients, err := is.ingredientRepo.CheckLowStockIngredients(ctx)
	if err != nil {
		return err
	}

	is.taskRepo.EnqueueAlertEmailTask(ctx, &repository.PayloadSendAlertEmail{Ingredients: ingredients})
	return nil
}
