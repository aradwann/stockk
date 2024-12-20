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
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"stockk/internal/config"
	"stockk/internal/controllers"
	"stockk/internal/db"
	"stockk/internal/repository"
	"stockk/internal/service"
)

func TestCreateOrderE2EWithTestcontainers(t *testing.T) {
	ctx := context.Background()

	// Spin up PostgreSQL container
	postgresContainer, dbConnString := setupPostgresContainer(t, ctx)
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate Postgres container: %v", err)
		}
	}()

	// Spin up Redis container
	redisContainer, redisAddress := setupRedisContainer(t, ctx)
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate Redis container: %v", err)
		}
	}()

	cfg := config.Config{
		DBDriver:      "pgx",
		DBSource:      dbConnString,
		RedisAddress:  redisAddress,
		MigrationsURL: "file://../db/migrations",
	}

	// Initialize database connection
	dbConn := db.InitDatabase(cfg)
	defer dbConn.Close()

	RedisClientOpts := asynq.RedisClientOpt{Addr: redisAddress}
	asynqClient := asynq.NewClient(RedisClientOpts)
	defer asynqClient.Close()

	// Initialize repositories
	ingredientRepo := repository.NewIngredientRepository(dbConn)
	orderRepo := repository.NewOrderRepository(dbConn)
	productRepo := repository.NewProductRepository(dbConn)
	taskQueueRepo := repository.NewTaskQueueRepository(asynqClient)

	// Initialize services
	orderService := service.NewOrderService(orderRepo, productRepo, ingredientRepo)
	ingredientService := service.NewIngredientService(ingredientRepo, taskQueueRepo)

	// Initialize controllers
	orderController := controllers.NewOrderController(orderService, ingredientService)

	// NOTE: database is already seeded in the migrations
	// with the following data:
	// - ingredients: Beef (20kg), Cheese (5kg), Onion (1kg)
	// - products: Burger (Beef 150g, Cheese 30g, Onion 20g)

	// Start test server
	router := setupRouter(orderController)
	server := httptest.NewServer(router)
	defer server.Close()

	// Define test payload
	payload := map[string]interface{}{
		"products": []map[string]interface{}{
			{"product_id": 1, "quantity": 2},
		},
	}
	requestBody, _ := json.Marshal(payload)

	// Perform HTTP POST request
	resp, err := http.Post(server.URL+"/api/v1/orders", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Assert response status
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code 201, got %d", resp.StatusCode)
	}

	// Assert database changes
	assertIngredientStockUpdated(t, dbConn, map[int]float64{
		1: 19700, // 20kg - 2 * 150g
		2: 4940,  // 5kg - 2 * 30g
		3: 960,   // 1kg - 2 * 20g
	})
}

// setupPostgresContainer sets up a PostgreSQL container.
func setupPostgresContainer(t *testing.T, ctx context.Context) (testcontainers.Container, string) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get PostgreSQL container host: %v", err)
	}
	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get PostgreSQL container port: %v", err)
	}

	connStr := "postgres://testuser:testpassword@" + host + ":" + port.Port() + "/testdb?sslmode=disable"
	return postgresContainer, connStr
}

// setupRedisContainer sets up a Redis container.
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
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get Redis container host: %v", err)
	}
	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("Failed to get Redis container port: %v", err)
	}

	addr := host + ":" + port.Port()
	return redisContainer, addr
}

// setupRouter sets up the HTTP router for testing.
func setupRouter(orderController *controllers.OrderController) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/v1/orders", orderController.CreateOrder)
	return r
}

// assertIngredientStockUpdated asserts that the ingredient stock levels are updated correctly.
func assertIngredientStockUpdated(t *testing.T, dbConn *sql.DB, expectedStock map[int]float64) {
	for id, expected := range expectedStock {
		var stock float64
		err := dbConn.QueryRow("SELECT current_stock FROM ingredients WHERE id = $1", id).Scan(&stock)
		if err != nil {
			t.Fatalf("Failed to fetch current_stock for ingredient %d: %v", id, err)
		}
		if stock != expected {
			t.Errorf("Expected current_stock for ingredient %d to be %f, got %f", id, expected, stock)
		}
	}
}
