package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/user/libreplex/calibre"    // Import the calibre package
	"github.com/user/libreplex/database" // Import the database package
)

const booksDir = "/books" // This is the directory mapped in Docker

var supportedExtensions = map[string]bool{
	".epub": true,
	".mobi": true,
	".azw3": true,
	// ".pdf": true, // Add PDF if calibre can reliably get metadata
}

func main() {
	// Initialize database connection using the default production path
	database.ConnectDefaultProductionDB()
	// The global DB instance can now be accessed via database.GetDB() if needed by handlers
	log.Println("Default production database initialized successfully.")

	r := gin.Default()

	// API v1 group
	apiV1 := r.Group("/api")
	{
		apiV1.GET("/health", func(c *gin.Context) {
			// You could add a DB ping here if desired, using database.GetDB()
			c.JSON(http.StatusOK, gin.H{
				"status":  "UP",
				"message": "ShelfStone is running",
			})
		})

		apiV1.POST("/ingest", handleIngest)
		apiV1.GET("/books", handleListBooks)
		apiV1.GET("/books/:id/download", handleDownloadBook)
	}

	// Static file serving for the frontend
	// Serve files from the ./frontend directory under the /ui path
	// e.g., /ui/index.html will serve ./frontend/index.html
	r.StaticFS("/ui", http.Dir("./frontend"))

	// Redirect /ui to /ui/index.html
	r.GET("/ui/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/ui/index.html")
	})


	r.Run(":8080") // Listen and serve on 0.0.0.0:8080
}

// Response structure for listing books
type BookResponse struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	AuthorName string `json:"authorName"`
	Format     string `json:"format"`
	FilePath   string `json:"filePath"` // May remove this from public API later for security
}

func handleListBooks(c *gin.Context) {
	db := database.GetDB()
	var books []database.Book
	// Eager load Author information.
	// .Preload("Author") tells GORM to also fetch the related Author for each Book.
	if err := db.Preload("Author").Order("title asc").Find(&books).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve books"})
		return
	}

	var response []BookResponse
	for _, b := range books {
		authorName := "Unknown"
		if b.Author.Name != "" { // Check if Author was successfully preloaded and has a name
			authorName = b.Author.Name
		}
		response = append(response, BookResponse{
			ID:         b.ID,
			Title:      b.Title,
			AuthorName: authorName,
			Format:     b.Format,
			FilePath:   b.FilePath, // Included for now, useful for prototype download link
		})
	}

	c.JSON(http.StatusOK, response)
}

func handleDownloadBook(c *gin.Context) {
	db := database.GetDB()
	bookID := c.Param("id")

	var book database.Book
	if err := db.First(&book, bookID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	// Ensure file path is not trying to escape the intended directory (basic check)
	// In a real scenario, ensure booksDir is an absolute path and resolved safely.
	// Here, book.FilePath is already absolute from the Docker container's perspective.
	// For added security, ensure it's under 'booksDir' if that's a concern.
	// However, since we store absolute paths from ingestion, this should be fine.

	// Check if file exists before attempting to serve
	if _, err := os.Stat(book.FilePath); os.IsNotExist(err) {
		log.Printf("File not found for book ID %s at path %s", bookID, book.FilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Book file not found on server"})
		return
	}

	// Set headers for file download
	// Use filepath.Base to get the filename from the path
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(book.FilePath))
	c.Header("Content-Type", "application/octet-stream") // General binary type
	// Could try to set a more specific MIME type based on book.Format if needed
	// e.g., c.Header("Content-Type", "application/epub+zip") for EPUB

	c.File(book.FilePath)
}

func handleIngest(c *gin.Context) {
	db := database.GetDB()
	var ingestedCount, skippedCount, errorCount int
	var ingestedFiles []string
	var errorMessages []string

	log.Println("Starting ingestion process...")

	err := filepath.Walk(booksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v. Skipping.", path, err)
			errorMessages = append(errorMessages, "Error accessing path "+path+": "+err.Error())
			errorCount++
			return nil // Continue walking even if one path is bad
		}

		if info.IsDir() {
			return nil // Skip directories
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !supportedExtensions[ext] {
			// log.Printf("Unsupported file type: %s. Skipping.", path)
			return nil // Skip unsupported file types
		}

		// Check if book already exists by FilePath
		var existingBook database.Book
		if err := db.Where("file_path = ?", path).First(&existingBook).Error; err == nil {
			// log.Printf("Book already exists in DB: %s. Skipping.", path)
			skippedCount++
			return nil
		}

		log.Printf("Processing new file: %s", path)

		// Get metadata using Calibre
		opfXML, err := calibre.GetBookMetadata(path)
		if err != nil {
			log.Printf("Error getting metadata for %s: %v", path, err)
			errorMessages = append(errorMessages, "Error getting metadata for "+path+": "+err.Error())
			errorCount++
			return nil // Continue with next file
		}

		metadata, err := calibre.ParseBookMetadataXML(opfXML)
		if err != nil {
			log.Printf("Error parsing metadata for %s: %v", path, err)
			errorMessages = append(errorMessages, "Error parsing metadata for "+path+": "+err.Error())
			errorCount++
			return nil // Continue with next file
		}

		// Find or create Author
		var author database.Author
		if metadata.Author != "" {
			// Using FirstOrCreate for Author
			// If author with this name is found, it's loaded into 'author'.
			// If not found, a new author with this name is created and loaded into 'author'.
			if authorRes := db.Where(database.Author{Name: metadata.Author}).FirstOrCreate(&author); authorRes.Error != nil {
				log.Printf("Error finding/creating author '%s' for book %s: %v", metadata.Author, path, authorRes.Error)
				errorMessages = append(errorMessages, "Error finding/creating author '"+metadata.Author+"': "+authorRes.Error.Error())
				errorCount++
				return nil // Continue with the next file
			}
		} else {
			// Handle cases where author might be empty or use a default/unknown author
			// For now, we'll assign to an "Unknown Author" or skip if that's not desired.
			// Let's create/use an "Unknown Author"
			if authorRes := db.Where(database.Author{Name: "Unknown Author"}).FirstOrCreate(&author); authorRes.Error != nil {
				log.Printf("Error finding/creating 'Unknown Author' for book %s: %v", path, authorRes.Error)
				errorMessages = append(errorMessages, "Error finding/creating 'Unknown Author': "+authorRes.Error.Error())
				errorCount++
				return nil
			}
		}

		book := database.Book{
			Title:         metadata.Title,
			AuthorID:      author.ID, // Assign the Author's ID
			// Author: author, // GORM will handle this relationship if AuthorID is set
			ISBN:          metadata.ISBN,
			FilePath:      path,
			Format:        strings.TrimPrefix(ext, "."),
			AddedAt:       time.Now(),
			Publisher:     metadata.Publisher,
			PublishedDate: metadata.PublishedDate,
			Description:   metadata.Description,
			Language:      metadata.Language,
		}

		if result := db.Create(&book); result.Error != nil {
			log.Printf("Error saving book %s to DB: %v", path, result.Error)
			errorMessages = append(errorMessages, "Error saving book "+path+" to DB: "+result.Error.Error())
			errorCount++
			return nil // Continue with next file
		}

		ingestedFiles = append(ingestedFiles, path)
		ingestedCount++
		log.Printf("Successfully ingested: %s", path)
		return nil
	})

	if err != nil {
		log.Printf("Error walking the path %s: %v", booksDir, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":       "Error during ingestion process",
			"error":         err.Error(),
			"ingestedCount": ingestedCount,
			"skippedCount":  skippedCount,
			"errorCount":    errorCount + 1, // Count this walking error as one
			"ingestedFiles": ingestedFiles,
			"errorMessages": append(errorMessages, "Error walking the books directory: "+err.Error()),
		})
		return
	}

	log.Println("Ingestion process completed.")
	c.JSON(http.StatusOK, gin.H{
		"message":       "Ingestion process completed",
		"ingestedCount": ingestedCount,
		"skippedCount":  skippedCount,
		"errorCount":    errorCount,
		"ingestedFiles": ingestedFiles,
		"errorMessages": errorMessages,
	})
}
