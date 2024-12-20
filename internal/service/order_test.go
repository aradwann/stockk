package service

import (
	"context"
	"stockk/internal/models"
	mockrepository "stockk/internal/repository/mock"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestCreateOrder(t *testing.T) {
	testCases := []struct {
		name       string
		input      []models.OrderItem
		buildStubs func(
			orderRepo *mockrepository.MockOrderRepository,
			productRepo *mockrepository.MockProductRepository,
			ingredientRepo *mockrepository.MockIngredientRepository,
			tx *mockrepository.MockTransaction,
		)
		buildContext func(t *testing.T) context.Context
		checkResult  func(t *testing.T, err error)
	}{
		{
			name: "Success Create Order",
			input: []models.OrderItem{
				{
					ProductID: 1,
					Quantity:  1,
				},
			},
			buildStubs: func(
				orderRepo *mockrepository.MockOrderRepository,
				productRepo *mockrepository.MockProductRepository,
				ingredientRepo *mockrepository.MockIngredientRepository,
				tx *mockrepository.MockTransaction,
			) {
				orderRepo.EXPECT().BeginTransaction().Return(tx, nil)
				orderRepo.EXPECT().CreateOrder(gomock.Any(), tx, gomock.Any()).Return(nil)

				productRepo.EXPECT().GetProductById(gomock.Any(), tx, gomock.Any()).
					Return(&models.Product{
						ID: 1,
						Ingredients: []models.ProductIngredient{
							{ProductID: 1, IngredientID: 1, Amount: 2},
						},
					}, nil)

				ingredientRepo.EXPECT().GetIngredientByID(gomock.Any(), tx, gomock.Any()).
					Return(&models.Ingredient{ID: 1, CurrentStock: 10}, nil)

				ingredientRepo.EXPECT().UpdateStock(gomock.Any(), tx, gomock.Any(), float64(8)).Return(nil)

				tx.EXPECT().Commit().Return(nil) // Expect commit on success
				tx.EXPECT().Rollback().Times(0)  // No rollback expected in success
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			orderRepo := mockrepository.NewMockOrderRepository(ctrl)
			productRepo := mockrepository.NewMockProductRepository(ctrl)
			ingredientRepo := mockrepository.NewMockIngredientRepository(ctrl)
			tx := mockrepository.NewMockTransaction(ctrl)

			tc.buildStubs(orderRepo, productRepo, ingredientRepo, tx)

			os := NewOrderService(orderRepo, productRepo, ingredientRepo)

			ctx := tc.buildContext(t)
			_, err := os.CreateOrder(ctx, tc.input)
			tc.checkResult(t, err)
		})
	}
}
