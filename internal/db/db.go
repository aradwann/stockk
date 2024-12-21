package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"stockk/internal/config"
)

// InitDatabase initializes the database connection and performs migrations.
func InitDatabase(config config.Config) *sql.DB {
	// Initialize database connection.
	dbConn, err := initDatabaseConn(config)
	if err != nil {
		handleError("Unable to connect to database", err)
	}

	// Run database migrations.
	RunDBMigrations(dbConn, config.MigrationsURL, config.UnversionedMigrationsUrl)

	return dbConn
}

// initDatabaseConn initializes the database connection.
func initDatabaseConn(config config.Config) (*sql.DB, error) {
	dbConn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		handleError("Unable to connect to database", err)
	}
	return dbConn, err
}

// handleError logs an error message and exits the program with status code 1.
func handleError(message string, err error) {
	slog.Error(fmt.Sprintf("%s: %v", message, err))
	os.Exit(1)
}
