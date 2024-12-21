package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"stockk/internal/config"
	"stockk/internal/controllers"
	"stockk/internal/db"
	"stockk/internal/repository"
	"stockk/internal/service"
)

func TestCreateOrderE2EWithTable(t *testing.T) {
	ctx := context.Background()

	// Set up PostgreSQL container
	postgresContainer := setupPostgresContainer(t, ctx)
	defer terminateContainer(t, postgresContainer, ctx)

	// Set up Redis container
	redisContainer, redisAddress := setupRedisContainer(t, ctx)
	defer terminateContainer(t, redisContainer, ctx)

	cfg := setupConfig(t, ctx, postgresContainer, redisAddress)
	dbConn := db.InitDatabase(cfg)
	defer dbConn.Close()

	asynqClient := setupAsynqClient(redisAddress)
	defer asynqClient.Close()

	ingredientRepo := repository.NewIngredientRepository(dbConn)
	orderRepo := repository.NewOrderRepository(dbConn)
	productRepo := repository.NewProductRepository(dbConn)
	taskQueueRepo := repository.NewTaskQueueRepository(asynqClient)

	orderService := service.NewOrderService(orderRepo, productRepo, ingredientRepo)
	ingredientService := service.NewIngredientService(ingredientRepo, taskQueueRepo)

	orderController := controllers.NewOrderController(orderService, ingredientService)

	router := setupRouter(orderController)
	server := httptest.NewServer(router)
	defer server.Close()

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedStock  map[int]float64
		expectedQueue  int
	}{
		{
			name: "Create order with valid stock",
			payload: map[string]interface{}{
				"products": []map[string]interface{}{
					{"product_id": 1, "quantity": 2},
				},
			},
			expectedStatus: http.StatusCreated,
			expectedStock: map[int]float64{
				1: 19700, // 20kg - 2 * 150g
				2: 4940,  // 5kg - 2 * 30g
				3: 960,   // 1kg - 2 * 20g
			},
			expectedQueue: 0,
		},
		{
			name: "Create large order triggering queue",
			payload: map[string]interface{}{
				"products": []map[string]interface{}{
					{"product_id": 1, "quantity": 40},
				},
			},
			expectedStatus: http.StatusCreated,
			expectedStock: map[int]float64{
				1: 13700, // 19700 - 40 * 150g
				2: 3740,  // 4940 - 40 * 30g
				3: 160,   // 960 - 40 * 20g
			},
			expectedQueue: 1,
		},
		{
			name: "invalid product id",
			payload: map[string]interface{}{
				"products": []map[string]interface{}{
					{"product_id": -1, "quantity": 40},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedStock:  map[int]float64{},
			expectedQueue:  0,
		},
		{
			name: "product not found",
			payload: map[string]interface{}{
				"products": []map[string]interface{}{
					{"product_id": 3, "quantity": 40},
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedStock:  map[int]float64{},
			expectedQueue:  0,
		},
		{
			name: "insufficient stock",
			payload: map[string]interface{}{
				"products": []map[string]interface{}{
					{"product_id": 1, "quantity": 999999},
				},
			},
			expectedStatus: http.StatusConflict,
			expectedStock:  map[int]float64{},
			expectedQueue:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, _ := json.Marshal(tt.payload)
			resp, err := http.Post(server.URL+"/api/v1/orders", "application/json", bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)
			assertIngredientStockUpdated(t, dbConn, tt.expectedStock)

			inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: redisAddress})
			defer inspector.Close()

			defaultQueue, err := inspector.GetQueueInfo("default")
			if tt.expectedQueue > 0 {
				require.NoError(t, err)
				require.Equal(t, tt.expectedQueue, defaultQueue.Size)
			}
		})
	}
}

func setupPostgresContainer(t *testing.T, ctx context.Context) *postgres.PostgresContainer {
	ctr, err := postgres.Run(
		ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("stock-test"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpassword"),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	require.NoError(t, err)
	return ctr
}

func setupRedisContainer(t *testing.T, ctx context.Context) (testcontainers.Container, string) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(60 * time.Second),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := redisContainer.Host(ctx)
	require.NoError(t, err)
	port, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	addr := host + ":" + port.Port()
	return redisContainer, addr
}

func terminateContainer(t *testing.T, container testcontainers.Container, ctx context.Context) {
	err := container.Terminate(ctx)
	require.NoError(t, err)
}

func setupConfig(t *testing.T, ctx context.Context, postgresContainer *postgres.PostgresContainer, redisAddress string) config.Config {
	dbURL, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)

	return config.Config{
		DBDriver:          "pgx",
		RedisAddress:      redisAddress,
		MigrationsURL:     "file://../db/migrations",
		TestMerchantEmail: "expected@domain.com",
		DBSource:          dbURL,
	}
}

func setupAsynqClient(redisAddress string) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddress})
}

func setupRouter(orderController *controllers.OrderController) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/v1/orders", orderController.CreateOrder)
	return r
}

func assertIngredientStockUpdated(t *testing.T, dbConn *sql.DB, expectedStock map[int]float64) {
	for id, expected := range expectedStock {
		var stock float64
		err := dbConn.QueryRow("SELECT current_stock FROM ingredients WHERE id = $1", id).Scan(&stock)
		require.NoError(t, err)
		require.Equal(t, expected, stock)
	}
}
