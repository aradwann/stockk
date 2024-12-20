MIGRATIONS_PATH=db/migrations

# Migration
migrateup:
	go run db/scripts/migrate.go
createmigration:
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq "$(filter-out $@,$(MAKECMDGOALS))"

# Mocks
mock:
	mockgen -package mockrepository -destination internal/repository/mock/repository.go stockk/internal/repository IngredientRepository,OrderRepository,ProductRepository,TaskQueueRepository,Transaction
	mockgen -package mockservice -destination internal/service/mock/service.go stockk/internal/service IngredientService,OrderService

# Testing
test: 
	go test -short -v -cover ./...

testci:
	go test -short -race -covermode atomic -coverprofile=covprofile ./...


.PHONY: migrateup createmigration mock test testci
