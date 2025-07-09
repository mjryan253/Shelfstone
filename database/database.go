package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDatabase initializes the database connection and migrates the schema.
// It expects a `config` directory to exist for storing the SQLite database file.
func ConnectDatabase() {
	var err error

	// Ensure the /config directory exists
	configDir := "./config"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		log.Printf("Config directory %s does not exist, creating it...\n", configDir)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Fatalf("Failed to create config directory: %v", err)
		}
	} else if err != nil {
		log.Fatalf("Error checking config directory: %v", err)
	}

	dbPath := filepath.Join(configDir, "libreplex.db")
	fmt.Printf("Database path: %s\n", dbPath)

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Log SQL queries
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established. Migrating schema...")

	// Auto-migrate the schema
	err = DB.AutoMigrate(&User{}, &Author{}, &Series{}, &Book{})
	if err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}

	log.Println("Database migration completed.")
}
