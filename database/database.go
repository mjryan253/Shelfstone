package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings" // Added import

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var globalDB *gorm.DB

// InitDB initializes the database connection using the given DSN and migrates the schema.
// It returns the GORM DB instance and an error if connection or migration fails.
func InitDB(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DSN cannot be empty")
	}

	// Ensure the directory for the DSN exists if it's a file-based SQLite DB
	// Adjusted condition: dsn is not empty here, and also not a special in-memory DSN
	if dsn != "file::memory:?cache=shared" && !strings.HasPrefix(dsn, "file:") && !strings.Contains(dsn, "@tcp(") { // A simple check for file paths vs DSNs like "user:pass@tcp(host:port)/dbname" or other non-file DSNs
		dbDir := filepath.Dir(dsn)
		if _, err := os.Stat(dbDir); os.IsNotExist(err) {
			log.Printf("Database directory %s does not exist, creating it...\n", dbDir)
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create database directory: %w", err)
			}
		} else if err != nil {
			return nil, fmt.Errorf("error checking database directory: %w", err)
		}
	}


	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Log SQL queries during info, or logger.Silent for tests
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database using DSN '%s': %w", dsn, err)
	}

	log.Println("Database connection established. Migrating schema...")

	// Auto-migrate the schema
	err = db.AutoMigrate(&User{}, &Author{}, &Series{}, &Book{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	log.Println("Database migration completed.")
	return db, nil
}

// ConnectDefaultProductionDB sets up the database connection for production using a default path
// and stores it in a global variable.
func ConnectDefaultProductionDB() {
	var err error
	configDir := "./config"
	if _, statErr := os.Stat(configDir); os.IsNotExist(statErr) {
		log.Printf("Config directory %s does not exist, creating it...\n", configDir)
		if mkdirErr := os.MkdirAll(configDir, 0755); mkdirErr != nil {
			log.Fatalf("Failed to create config directory: %v", mkdirErr)
		}
	} else if statErr != nil {
		log.Fatalf("Error checking config directory: %v", statErr)
	}

	dbPath := filepath.Join(configDir, "libreplex.db")
	log.Printf("Default production database path: %s\n", dbPath)

	globalDB, err = InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize default production database: %v", err)
	}
}

// GetDB returns the global DB instance.
// This is useful for parts of the application that rely on a global DB instance.
// Ensure ConnectDefaultProductionDB has been called before using this.
func GetDB() *gorm.DB {
	if globalDB == nil {
		log.Println("Warning: GetDB called before ConnectDefaultProductionDB or connection failed. Attempting to connect now.")
		// Attempt a fallback connection. This might not be ideal for all scenarios.
		ConnectDefaultProductionDB()
		if globalDB == nil { // If still nil after attempt
			log.Fatal("Failed to establish database connection on fallback.")
		}
	}
	return globalDB
}
