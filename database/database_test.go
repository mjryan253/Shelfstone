package database

import (
	"errors" // Added import
	"testing"

	"gorm.io/gorm"
)

// TestInitDB_InMemory tests database initialization with an in-memory SQLite database.
func TestInitDB_InMemory(t *testing.T) {
	// Use "file::memory:?cache=shared" for an in-memory SQLite database for testing
	dsn := "file::memory:?cache=shared"
	// dsn := filepath.Join(t.TempDir(), "test.db") // Alternative: use a temp file DB

	// Call the package-level InitDB function
	db, err := InitDB(dsn)
	if err != nil {
		t.Fatalf("database.InitDB failed: %v", err)
	}

	// Check if GORM instance is returned
	if db == nil {
		t.Fatal("database.InitDB returned a nil DB instance")
	}

	// Verify that tables for the models were created
	// These are the table names GORM typically creates (pluralized snake_case)
	expectedTables := []string{"users", "authors", "series", "books"}

	for _, tableName := range expectedTables {
		if !db.Migrator().HasTable(tableName) {
			// Check with the model struct name if the pluralized version fails
			// This can happen based on GORM's table naming strategy or specific model annotations
			var modelInstance interface{}
			switch tableName {
			case "users":
				modelInstance = &User{}
			case "authors":
				modelInstance = &Author{}
			case "series":
				modelInstance = &Series{}
			case "books":
				modelInstance = &Book{}
			default:
				t.Errorf("Unknown table name %s in test logic", tableName)
				continue
			}
			if !db.Migrator().HasTable(modelInstance) {
				t.Errorf("Table '%s' (or for model '%T') was not created by InitDB migrations", tableName, modelInstance)
			}
		}
	}

	// Try a simple query to ensure the database is responsive (optional)
	var userCount int64
	err = db.Model(&User{}).Count(&userCount).Error
	if err != nil {
		t.Errorf("Failed to query database after InitDB: %v", err)
	}
	if userCount != 0 {
		t.Errorf("Expected 0 users in a fresh database, got %d", userCount)
	}

	// You can get the underlying SQL DB and close it if using a file-based temp DB
	sqlDB, err := db.DB()
	if err != nil {
		t.Logf("Error getting underlying *sql.DB: %v", err)
	} else {
		err = sqlDB.Close()
		if err != nil {
			t.Logf("Error closing test database: %v", err)
		}
	}
}

// TestInitDB_EmptyDSN tests InitDB with an empty DSN, expecting an error.
func TestInitDB_EmptyDSN(t *testing.T) {
	_, err := InitDB("") // Calls package-level InitDB
	if err == nil {
		t.Errorf("Expected an error when calling database.InitDB with an empty DSN, but got nil")
	}
}

// TestCRUDOperations (Optional, but good for testing Phase 1 DB functionality)
// This test is more involved and would test creating, reading, updating, and deleting records for each model.
// For Phase 1, ensuring InitDB runs and creates schema is the primary goal.
// We can expand this later.
func TestBookAuthorSeriesCRUD(t *testing.T) {
	dsn := "file::memory:?cache=shared"
	db, err := InitDB(dsn) // Calls package-level InitDB
	if err != nil {
		t.Fatalf("database.InitDB failed: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Create Author
	author := Author{Name: "Test Author CRUD"}
	result := db.Create(&author)
	if result.Error != nil {
		t.Fatalf("Failed to create author: %v", result.Error)
	}
	if author.ID == 0 {
		t.Fatalf("Author ID not set after creation")
	}

	// Create Series
	series := Series{Name: "Test Series CRUD"}
	result = db.Create(&series)
	if result.Error != nil {
		t.Fatalf("Failed to create series: %v", result.Error)
	}
	if series.ID == 0 {
		t.Fatalf("Series ID not set after creation")
	}

	// Create Book
	book := Book{
		Title:    "Test Book CRUD",
		AuthorID: author.ID,
		SeriesID: &series.ID,
		FilePath: "/test/book_crud.epub",
		Format:   "EPUB",
	}
	result = db.Create(&book)
	if result.Error != nil {
		t.Fatalf("Failed to create book: %v", result.Error)
	}
	if book.ID == 0 {
		t.Fatalf("Book ID not set after creation")
	}

	// Read Book (and implicitly its relations)
	var fetchedBook Book
	// Preload Author and Series to test relationships
	err = db.Preload("Author").Preload("Series").First(&fetchedBook, book.ID).Error
	if err != nil {
		t.Fatalf("Failed to fetch book: %v", err)
	}

	if fetchedBook.Title != "Test Book CRUD" {
		t.Errorf("Fetched book title mismatch: expected 'Test Book CRUD', got '%s'", fetchedBook.Title)
	}
	if fetchedBook.Author.Name != "Test Author CRUD" {
		t.Errorf("Fetched book author name mismatch: expected 'Test Author CRUD', got '%s'", fetchedBook.Author.Name)
	}
	if fetchedBook.Series == nil || fetchedBook.Series.Name != "Test Series CRUD" {
		t.Errorf("Fetched book series name mismatch: expected 'Test Series CRUD', got '%s'", fetchedBook.Series.Name)
	}

	// Update Book
	updatedTitle := "Updated Test Book CRUD"
	result = db.Model(&fetchedBook).Update("Title", updatedTitle)
	if result.Error != nil {
		t.Fatalf("Failed to update book: %v", result.Error)
	}
	if fetchedBook.Title != updatedTitle {
		t.Errorf("Book title not updated in struct after Model().Update(): expected '%s', got '%s'", updatedTitle, fetchedBook.Title)
		// Re-fetch to confirm DB state
		var refetchedBook Book
		db.First(&refetchedBook, book.ID)
		if refetchedBook.Title != updatedTitle {
			t.Errorf("Book title not updated in DB: expected '%s', got '%s'", updatedTitle, refetchedBook.Title)
		}
	}


	// Delete Book
	result = db.Delete(&fetchedBook)
	if result.Error != nil {
		t.Fatalf("Failed to delete book: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		t.Errorf("No rows affected when deleting book.")
	}

	// Verify Book is deleted
	var checkBook Book
	err = db.First(&checkBook, book.ID).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("Expected ErrRecordNotFound after deleting book, but got: %v", err)
	}

	// Delete Author and Series (cleanup)
	db.Delete(&author)
	db.Delete(&series)
}
