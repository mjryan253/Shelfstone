package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

const coversDir = "/cache/covers" // Directory to store extracted cover images

func main() {
	// Ensure covers directory exists
	if err := os.MkdirAll(coversDir, 0755); err != nil {
		log.Fatalf("Failed to create covers directory %s: %v", coversDir, err)
	}
	log.Printf("Covers directory ensured at: %s", coversDir)

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
		apiV1.GET("/books/:id", handleGetBookDetails) // New endpoint for single book metadata
		apiV1.GET("/books/:id/download", handleDownloadBook)
		// Endpoint to serve cover images
		// Example: /api/covers/book_1_cover.jpg
		// The filename will be stored in Book.CoverPath (relative to coversDir or an identifier)
		// For simplicity, let's assume CoverPath stores just the filename like "book_1_cover.jpg"
		// and we serve it from a fixed directory.
		apiV1.StaticFS("/covers", http.Dir(coversDir))
	}

	// Static file serving for the frontend
	// Serve files from the ./frontend directory under the /ui path
	// e.g., /ui/index.html will serve ./frontend/index.html
	r.StaticFS("/ui", http.Dir("./frontend"))

	// Redirect /ui to /ui/index.html
	r.GET("/ui/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/ui/index.html")
	})


	// Start periodic library scanner
	go startLibraryScanner(5 * time.Minute) // Scan every 5 minutes

	r.Run(":8080") // Listen and serve on 0.0.0.0:8080
}

// processFile is a helper function to process a single book file.
// It encapsulates the logic from the filepath.WalkFunc in handleIngest.
// Returns true if ingested, false if skipped/error, and any error message.
func processFile(path string, info os.FileInfo) (ingested bool, skipped bool, errorMessage string) {
	db := database.GetDB()

	if info.IsDir() {
		return false, true, "" // Skip directories
	}

	ext := strings.ToLower(filepath.Ext(path))
	if !supportedExtensions[ext] {
		return false, true, "" // Skip unsupported file types
	}

	// Check if book already exists by FilePath
	var existingBook database.Book
	if err := db.Where("file_path = ?", path).First(&existingBook).Error; err == nil {
		// log.Printf("Book already exists in DB: %s. Skipping.", path)
		return false, true, "" // Already exists, skipped
	}

	log.Printf("Processing new file: %s", path)

	// Get metadata using Calibre
	opfXML, err := calibre.GetBookMetadata(path)
	if err != nil {
		log.Printf("Error getting metadata for %s: %v", path, err)
		return false, false, fmt.Sprintf("Error getting metadata for %s: %v", path, err)
	}

	metadata, err := calibre.ParseBookMetadataXML(opfXML)
	if err != nil {
		log.Printf("Error parsing metadata for %s: %v", path, err)
		return false, false, fmt.Sprintf("Error parsing metadata for %s: %v", path, err)
	}

	// Find or create Author
	var author database.Author
	if metadata.Author != "" {
		if authorRes := db.Where(database.Author{Name: metadata.Author}).FirstOrCreate(&author); authorRes.Error != nil {
			log.Printf("Error finding/creating author '%s' for book %s: %v", metadata.Author, path, authorRes.Error)
			return false, false, fmt.Sprintf("Error finding/creating author '%s': %v", metadata.Author, authorRes.Error)
		}
	} else {
		if authorRes := db.Where(database.Author{Name: "Unknown Author"}).FirstOrCreate(&author); authorRes.Error != nil {
			log.Printf("Error finding/creating 'Unknown Author' for book %s: %v", path, authorRes.Error)
			return false, false, fmt.Sprintf("Error finding/creating 'Unknown Author': %v", authorRes.Error)
		}
	}

	// Create the book record first to get an ID for the cover filename,
	// and to link series. We will update it with the cover path later.
	tempBook := database.Book{
		Title:         metadata.Title,
		AuthorID:      author.ID,
		ISBN:          metadata.ISBN,
		FilePath:      path,
		Format:        strings.TrimPrefix(ext, "."),
		AddedAt:       time.Now(),
		Publisher:     metadata.Publisher,
		PublishedDate: metadata.PublishedDate,
		Description:   metadata.Description,
		Language:      metadata.Language,
	}

	// Handle Series information
	if metadata.Series != "" {
		var series database.Series
		if seriesRes := db.Where(database.Series{Name: metadata.Series}).FirstOrCreate(&series); seriesRes.Error != nil {
			log.Printf("Error finding/creating series '%s' for book %s: %v", metadata.Series, path, seriesRes.Error)
			// Decide if this is a fatal error for the book or just a warning
			// For now, log and continue without series info
		} else {
			tempBook.SeriesID = &series.ID
			if metadata.SeriesIndex != "" {
				seriesIndexVal, err := strconv.ParseFloat(metadata.SeriesIndex, 64)
				if err != nil {
					log.Printf("Error parsing series index '%s' for book %s: %v", metadata.SeriesIndex, path, err)
				} else {
					tempBook.SeriesIndex = &seriesIndexVal
				}
			}
		}
	}

	if result := db.Create(&tempBook); result.Error != nil {
		log.Printf("Error saving initial book record for %s to DB: %v", path, result.Error)
		return false, false, fmt.Sprintf("Error saving initial book record for %s to DB: %v", path, result.Error)
	}

	// Attempt to extract cover image
	// Use book ID to create a unique cover filename, e.g., cover_123.jpg
	// Calibre's --get-cover usually outputs as .jpg by default if not specified.
	// Let's assume .jpg for now.
	coverFilename := fmt.Sprintf("cover_%d.jpg", tempBook.ID)
	outputCoverPath := filepath.Join(coversDir, coverFilename)

	err = calibre.ExtractCoverImage(path, outputCoverPath)
	if err != nil {
		if strings.Contains(err.Error(), "No cover found") {
			log.Printf("No cover image found for book ID %d (%s). Continuing without cover.", tempBook.ID, path)
		} else {
			log.Printf("Error extracting cover image for book ID %d (%s): %v. Continuing without cover.", tempBook.ID, path, err)
			// We could choose to add this to errorMessages and increment errorCount for scanLibrary
			// For now, logging it and proceeding without a cover.
		}
		tempBook.CoverPath = "" // Ensure it's empty or null if no cover
	} else {
		log.Printf("Successfully extracted cover image for book ID %d to %s", tempBook.ID, outputCoverPath)
		// Store the relative path or just the filename to be served via /api/covers/
		// Storing just the filename is simpler for the StaticFS setup.
		tempBook.CoverPath = coverFilename
	}

	// Update book record with cover path
	if err := db.Save(&tempBook).Error; err != nil {
		log.Printf("Error updating book record ID %d with cover path: %v", tempBook.ID, err)
		// This is an error, but the book is already created. We might want to surface this.
		// For now, log and continue. The ingestion of the book itself was successful.
		// We could potentially add this to the overall errorMessages for the scan.
		// Let's return it as part of errorMessage for this file processing.
		return false, false, fmt.Sprintf("Error updating book ID %d with cover path: %v", tempBook.ID, err)
	}


	log.Printf("Successfully ingested and processed cover for: %s (Book ID: %d)", path, tempBook.ID)
	return true, false, "" // Ingested successfully
}

// scanLibrary performs a one-time scan of the booksDir.
func scanLibrary() (ingestedCount int, skippedCount int, errorCount int, errorMessages []string) {
	log.Println("Periodic library scan initiated...")
	err := filepath.Walk(booksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v. Skipping.", path, err)
			errorMessages = append(errorMessages, fmt.Sprintf("Error accessing path %s: %v", path, err))
			errorCount++
			return nil // Continue walking
		}

		// Skip the root booksDir itself if it's passed directly to filepath.Walk
		// (info for booksDir will be a directory, so processFile will skip it anyway, but good to be explicit)
		if path == booksDir && info.IsDir() {
			return nil
		}


		wasIngested, wasSkipped, errMsg := processFile(path, info)
		if errMsg != "" {
			errorMessages = append(errorMessages, errMsg)
			errorCount++
		} else if wasIngested {
			ingestedCount++
		} else if wasSkipped {
			skippedCount++
		}
		return nil // Continue walking
	})

	if err != nil {
		log.Printf("Error walking the path %s during periodic scan: %v", booksDir, err)
		errorMessages = append(errorMessages, fmt.Sprintf("Error walking the books directory: %v", err))
		errorCount++ // Count this walking error
	}

	if ingestedCount > 0 || errorCount > 0 { // Log only if something happened or errors occurred
		log.Printf("Periodic library scan finished. Ingested: %d, Skipped: %d, Errors: %d.", ingestedCount, skippedCount, errorCount)
		if errorCount > 0 {
			for _, emsg := range errorMessages {
				log.Printf("Scan Error: %s", emsg)
			}
		}
	} else {
		log.Println("Periodic library scan finished. No new files found or errors.")
	}
	return
}

// startLibraryScanner initiates periodic scanning of the library.
func startLibraryScanner(interval time.Duration) {
	log.Printf("Library scanner starting: initial scan, then every %v.", interval)

	// Perform an initial scan immediately on startup
	scanLibrary()

	// Then, scan periodically
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		scanLibrary()
	}
}

// Response structure for listing books
type BookResponse struct {
	ID          uint     `json:"id"`
	Title       string   `json:"title"`
	AuthorName  string   `json:"authorName"`
	Format      string   `json:"format"`
	FilePath    string   `json:"filePath"` // May remove this from public API later for security
	CoverURL    string   `json:"coverUrl,omitempty"` // URL to the cover image
	SeriesName  string   `json:"seriesName,omitempty"`
	SeriesIndex *float64 `json:"seriesIndex,omitempty"`
}

func handleListBooks(c *gin.Context) {
	db := database.GetDB()
	var books []database.Book
	// Eager load Author and Series information.
	if err := db.Preload("Author").Preload("Series").Order("title asc").Find(&books).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve books"})
		return
	}

	var response []BookResponse
	for _, b := range books {
		authorName := "Unknown"
		if b.Author.ID != 0 && b.Author.Name != "" { // Check if Author was successfully preloaded
			authorName = b.Author.Name
		}

		var coverURL string
		if b.CoverPath != "" {
			coverURL = fmt.Sprintf("/api/covers/%s", b.CoverPath)
		}

		var seriesName string
		if b.SeriesID != nil && b.Series != nil && b.Series.ID != 0 { // Check if Series was successfully preloaded
			seriesName = b.Series.Name
		}

		response = append(response, BookResponse{
			ID:          b.ID,
			Title:       b.Title,
			AuthorName:  authorName,
			Format:      b.Format,
			FilePath:    b.FilePath,
			CoverURL:    coverURL,
			SeriesName:  seriesName,
			SeriesIndex: b.SeriesIndex, // This is already a *float64, so it's omitempty if nil
		})
	}

	c.JSON(http.StatusOK, response)
}

func handleGetBookDetails(c *gin.Context) {
	db := database.GetDB()
	bookID := c.Param("id")

	var book database.Book
	// Eager load Author and Series information for the single book
	if err := db.Preload("Author").Preload("Series").First(&book, bookID).Error; err != nil {
		// Could be gorm.ErrRecordNotFound, handle appropriately
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	authorName := "Unknown"
	if book.Author.ID != 0 && book.Author.Name != "" {
		authorName = book.Author.Name
	}

	var coverURL string
	if book.CoverPath != "" {
		coverURL = fmt.Sprintf("/api/covers/%s", book.CoverPath)
	}

	var seriesName string
	if book.SeriesID != nil && book.Series != nil && book.Series.ID != 0 {
		seriesName = book.Series.Name
	}

	response := BookResponse{
		ID:          book.ID,
		Title:       book.Title,
		AuthorName:  authorName,
		Format:      book.Format,
		FilePath:    book.FilePath, // Consider if this should be exposed here
		CoverURL:    coverURL,
		SeriesName:  seriesName,
		SeriesIndex: book.SeriesIndex,
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
	log.Println("Manual ingestion process triggered via API.")
	// Note: The `scanLibrary` function now returns ingestedFiles and errorMessages.
	// However, for the API response, we might not need `ingestedFiles` if it's too verbose.
	// The counts and generic error messages should suffice.
	// Let's adapt the response to be consistent.
	// For now, scanLibrary doesn't return ingestedFiles list, only counts and error messages.
	// We can add it if desired, but it might make logs very long for periodic scans.

	ingestedCount, skippedCount, errorCount, errorMessages := scanLibrary()

	// The old handleIngest had a concept of "ingestedFiles" list.
	// scanLibrary currently doesn't collect this to keep its logging concise for periodic runs.
	// If we want to return it for the manual ingest API, we'd need to modify scanLibrary
	// or have processFile return the path on success for collection.
	// For now, let's omit `ingestedFiles` from the JSON response to keep it simple.

	if errorCount > 0 && len(errorMessages) > 0 && strings.Contains(errorMessages[len(errorMessages)-1], "Error walking the books directory") {
		// This indicates a fundamental issue with walking the directory, which is a server error.
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":       "Error during ingestion walk process",
			"ingestedCount": ingestedCount,
			"skippedCount":  skippedCount,
			"errorCount":    errorCount,
			"errorMessages": errorMessages,
		})
		return
	}

	log.Printf("Manual ingestion process finished. Ingested: %d, Skipped: %d, Errors: %d.", ingestedCount, skippedCount, errorCount)
	c.JSON(http.StatusOK, gin.H{
		"message":       "Ingestion process completed",
		"ingestedCount": ingestedCount,
		"skippedCount":  skippedCount,
		"errorCount":    errorCount,
		"errorMessages": errorMessages, // Contains detailed error messages from processing individual files
	})
}
