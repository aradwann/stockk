MIGRATIONS_PATH=db/migrations

# Migration
migrateup:
	go run db/scripts/migrate.go
createmigration:
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq "$(filter-out $@,$(MAKECMDGOALS))"


.PHONY: migrateup createmigration
