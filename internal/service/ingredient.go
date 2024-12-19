package service

import (
	"context"
	"errors"
	internalErrors "stockk/internal/errors"
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
			if errors.Is(err, internalErrors.ErrNotFound) {
				return internalErrors.Wrap(internalErrors.ErrNotFound, "ingredient service: ingredient not found")
			} else {
				return internalErrors.Wrap(internalErrors.ErrInternalServer, "ingredient service: failed to update ingredient")
			}

		}
	}
	return nil
}

func (is *IngredientService) CheckIngredientLevelsAndAlert(ctx context.Context) error {
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
