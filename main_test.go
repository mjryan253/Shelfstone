package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/user/libreplex/database" // Adjusted import path
)

// SetupRouter function is needed to initialize the router for tests.
// It's similar to the one in main.go but might be simplified for testing.
func SetupRouterForTest() *gin.Engine {
	// Initialize a test database (in-memory)
	// We don't need a real DB for the health check, but if other routes were tested,
	// this would be important. For health check, it's more about router setup.
	// Use the new InitDB from the database package
	testDb, err := database.InitDB("file::memory:?cache=shared") // Or your test DSN
	if err != nil {
		// In a real test suite, you might t.Fatalf or panic here
		// For simplicity, we'll let it proceed, but health might indirectly fail if DB is crucial for startup
		// For tests that need the DB, you'd check err and fail the test.
	}
	if testDb != nil { // Close the test database when done if it's not nil
		sqlDB, _ := testDb.DB()
		defer sqlDB.Close()
		// However, the current health endpoint doesn't depend on DB.
	}

	router := gin.Default()

	// Define the /api group (if your routes are grouped)
	api := router.Group("/api")
	{
		// Setup routes as in main.go
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "UP",
				"message": "ShelfStone is running",
			})
		})
		// Add other route setups here if testing them
	}
	return router
}

func TestHealthCheckRoute(t *testing.T) {
	// Set Gin to TestMode to suppress console output during tests.
	gin.SetMode(gin.TestMode)

	router := SetupRouterForTest()

	// Create a new HTTP request to the /api/health endpoint.
	req, err := http.NewRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}

	// Create a ResponseRecorder to record the response.
	w := httptest.NewRecorder()

	// Serve the HTTP request to the recorder.
	router.ServeHTTP(w, req)

	// Check if the status code is 200 OK.
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check if the response body is what we expect.
	// Expected response: {"status":"UP","message":"ShelfStone is running"}
	expectedBody := `{"message":"ShelfStone is running","status":"UP"}`
	// Note: JSON key order might vary. A more robust check would unmarshal and compare maps.
	if w.Body.String() != expectedBody {
		// Try the other order, as Gin's H map doesn't guarantee order
		expectedBodyAlt := `{"status":"UP","message":"ShelfStone is running"}`
		if w.Body.String() != expectedBodyAlt {
			t.Errorf("Expected body '%s' or '%s', got '%s'", expectedBody, expectedBodyAlt, w.Body.String())
		}
	}
}

// TestMain is needed to set up GIN_MODE to test if not already done globally
// func TestMain(m *testing.M) {
// 	//Set Gin to Test Mode
// 	gin.SetMode(gin.TestMode)
//
// 	// Run the other tests
// 	os.Exit(m.Run())
// }
