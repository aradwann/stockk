package service

import (
	"context"
	"errors"
	internalErrors "stockk/internal/errors"
	"stockk/internal/models"
	mockrepository "stockk/internal/repository/mock"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestUpdateIngredientStock(t *testing.T) {

	testCases := []struct {
		name         string
		input        []models.Ingredient
		buildStubs   func(ingredientrepo *mockrepository.MockIngredientRepository, taskRepo *mockrepository.MockTaskQueueRepository)
		buildContext func(t *testing.T) context.Context
		checkResult  func(t *testing.T, err error)
	}{
		{
			name: "Success Update",
			input: []models.Ingredient{
				{
					ID:           1,
					Name:         "Sugar",
					TotalStock:   100,
					CurrentStock: 40,
					AlertSent:    false,
				},
			},
			buildStubs: func(ingredientrepo *mockrepository.MockIngredientRepository, taskRepo *mockrepository.MockTaskQueueRepository) {
				ingredientrepo.EXPECT().UpdateStock(gomock.Any(), nil, 1, float64(40)).Return(nil)
			},
			buildContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			checkResult: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			},
		},
		{
			name: "Error Update",
			input: []models.Ingredient{
				{
					ID:           1,
					Name:         "Sugar",
					TotalStock:   100,
					CurrentStock: 40,
					AlertSent:    false,
				},
			},
			buildStubs: func(ingredientrepo *mockrepository.MockIngredientRepository, taskRepo *mockrepository.MockTaskQueueRepository) {
				ingredientrepo.EXPECT().UpdateStock(gomock.Any(), nil, 1, float64(40)).Return(errors.New("error"))
			},
			buildContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			checkResult: func(t *testing.T, err error) {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			},
		},
		{
			name: "Error Not Found",
			input: []models.Ingredient{
				{
					ID:           1,
					Name:         "Sugar",
					TotalStock:   100,
					CurrentStock: 40,
					AlertSent:    false,
				},
			},
			buildStubs: func(ingredientrepo *mockrepository.MockIngredientRepository, taskRepo *mockrepository.MockTaskQueueRepository) {
				ingredientrepo.EXPECT().UpdateStock(gomock.Any(), nil, 1, float64(40)).Return(internalErrors.ErrNotFound)
			},
			buildContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			checkResult: func(t *testing.T, err error) {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ir := mockrepository.NewMockIngredientRepository(ctrl)
			tr := mockrepository.NewMockTaskQueueRepository(ctrl)

			tc.buildStubs(ir, tr)
			is := NewIngredientService(ir, tr)

			ctx := tc.buildContext(t)

			err := is.UpdateIngredientStock(ctx, tc.input)
			tc.checkResult(t, err)
		})
	}
}

func TestCheckIngredientLevelsAndAlert(t *testing.T) {

	testCases := []struct {
		name         string
		buildStubs   func(ingredientrepo *mockrepository.MockIngredientRepository, taskRepo *mockrepository.MockTaskQueueRepository)
		buildContext func(t *testing.T) context.Context
		checkResult  func(t *testing.T, err error)
	}{
		{
			name: "Success Check",
			buildStubs: func(ingredientrepo *mockrepository.MockIngredientRepository, taskRepo *mockrepository.MockTaskQueueRepository) {
				ingredientrepo.EXPECT().CheckLowStockIngredients(gomock.Any()).Return([]models.Ingredient{}, nil)
			},
			buildContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			checkResult: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			},
		},
		{
			name: "Error Check",
			buildStubs: func(ingredientrepo *mockrepository.MockIngredientRepository, taskRepo *mockrepository.MockTaskQueueRepository) {
				ingredientrepo.EXPECT().CheckLowStockIngredients(gomock.Any()).Return(nil, errors.New("error"))
			},
			buildContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			checkResult: func(t *testing.T, err error) {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			},
		},
		{
			name: "Success Check and Enqueue",
			buildStubs: func(ingredientrepo *mockrepository.MockIngredientRepository, taskRepo *mockrepository.MockTaskQueueRepository) {
				ingredientrepo.EXPECT().CheckLowStockIngredients(gomock.Any()).Return([]models.Ingredient{{ID: 1}}, nil)
				taskRepo.EXPECT().EnqueueAlertEmailTask(gomock.Any(), gomock.Any()).Return(nil)
			},
			buildContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			checkResult: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			},
		},
		{
			name: "Error Check and Enqueue",
			buildStubs: func(ingredientrepo *mockrepository.MockIngredientRepository, taskRepo *mockrepository.MockTaskQueueRepository) {
				ingredientrepo.EXPECT().CheckLowStockIngredients(gomock.Any()).Return([]models.Ingredient{{ID: 1}}, nil)
				taskRepo.EXPECT().EnqueueAlertEmailTask(gomock.Any(), gomock.Any()).Return(errors.New("error"))
			},
			buildContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			checkResult: func(t *testing.T, err error) {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ir := mockrepository.NewMockIngredientRepository(ctrl)
			tr := mockrepository.NewMockTaskQueueRepository(ctrl)

			tc.buildStubs(ir, tr)
			is := NewIngredientService(ir, tr)

			ctx := tc.buildContext(t)

			err := is.CheckIngredientLevelsAndAlert(ctx)
			tc.checkResult(t, err)
		})
	}
}
