package service

import (
	"context"
	internalErrors "stockk/internal/errors"
	"stockk/internal/models"
	"stockk/internal/repository"
)

type IngredientService interface {
	UpdateIngredientStock(ctx context.Context, ingredients []models.Ingredient) error
	CheckIngredientLevelsAndAlert(ctx context.Context) error
}
type ingredientService struct {
	ingredientRepo repository.IngredientRepository
	taskRepo       repository.TaskQueueRepository
}

func NewIngredientService(ingredientRepo repository.IngredientRepository, taskRepo repository.TaskQueueRepository) IngredientService {
	return &ingredientService{ingredientRepo: ingredientRepo, taskRepo: taskRepo}
}

var _ IngredientService = (*ingredientService)(nil)

func (is *ingredientService) UpdateIngredientStock(ctx context.Context, ingredients []models.Ingredient) error {
	for _, ingredient := range ingredients {
		// Update stock in database
		if err := is.ingredientRepo.UpdateStock(ctx, nil, ingredient.ID, ingredient.CurrentStock); err != nil {
			return err
		}
	}
	return nil
}

func (is *ingredientService) CheckIngredientLevelsAndAlert(ctx context.Context) error {
	ingredients, err := is.ingredientRepo.CheckLowStockIngredients(ctx)
	if err != nil {
		return internalErrors.Wrap(internalErrors.ErrInternalServer, "ingredient service: failed to retrieve low stock ingredients")
	}

	if len(ingredients) > 0 {
		if err := is.taskRepo.EnqueueAlertEmailTask(ctx, &repository.PayloadSendAlertEmail{Ingredients: ingredients}); err != nil {
			return internalErrors.Wrap(internalErrors.ErrInternalServer, "ingredient service: failed to enqueue alert email task")

		}
	}
	return nil
}
