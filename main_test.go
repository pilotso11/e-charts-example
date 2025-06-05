package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupTestRouter creates a gin router for testing by setting the test mode
// and then using the common setupRouter function
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return setupRouter()
}

// TestGenerateChart tests the generateChart function
func TestGenerateChart(t *testing.T) {
	// Set a fixed seed for deterministic random numbers
	rand.Seed(42)

	// Create a valid chart request
	req := ChartRequest{
		From:          "2023-01-01",
		To:            "2023-01-31",
		Metric:        "api_to_or",
		GroupBy:       []string{"exchange_id", "account_id"},
		GroupFilter:   []string{"filter1", "filter2"},
		CentileFilter: []string{"p50", "p95"},
	}

	// Generate chart
	html, err := generateChart(req)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, html)
	assert.Contains(t, html, "api_to_or") // Title should contain metric
	assert.Contains(t, html, "From 2023-01-01 To 2023-01-31") // Subtitle should contain date range
	assert.Contains(t, html, "exchange_id") // Should contain group by values
	assert.Contains(t, html, "account_id")
}

// TestGenerateChartEmptyGroupBy tests the generateChart function with empty GroupBy
func TestGenerateChartEmptyGroupBy(t *testing.T) {
	// Set a fixed seed for deterministic random numbers
	rand.Seed(42)

	// Create a chart request with empty GroupBy
	req := ChartRequest{
		From:          "2023-01-01",
		To:            "2023-01-31",
		Metric:        "api_to_or",
		GroupBy:       []string{},
		GroupFilter:   []string{},
		CentileFilter: []string{},
	}

	// Generate chart
	html, err := generateChart(req)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, html)
	assert.Contains(t, html, "Default") // Should contain default series
}

// TestRootEndpoint tests the root endpoint
func TestRootEndpoint(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "Performance Reports") // Title from embedded HTML
}

// TestChartEndpoint tests the /chart endpoint with valid request
func TestChartEndpoint(t *testing.T) {
	router := setupTestRouter()

	// Create a valid chart request
	req := ChartRequest{
		From:          "2023-01-01",
		To:            "2023-01-31",
		Metric:        "api_to_or",
		GroupBy:       []string{"exchange_id"},
		GroupFilter:   []string{},
		CentileFilter: []string{},
	}

	// Convert to JSON
	jsonData, _ := json.Marshal(req)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/chart", bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "api_to_or") // Chart title should contain metric
}

// TestChartEndpointInvalidRequest tests the /chart endpoint with invalid request
func TestChartEndpointInvalidRequest(t *testing.T) {
	router := setupTestRouter()

	// Invalid JSON
	invalidJSON := `{"from": "2023-01-01", "to": "2023-01-31", "metric": "api_to_or", "groupBy": [}`

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/chart", strings.NewReader(invalidJSON))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Parse response
	var response ChartResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "Invalid request body", response.Message)
}

// TestChartEndpointMissingFields tests the /chart endpoint with missing required fields
func TestChartEndpointMissingFields(t *testing.T) {
	router := setupTestRouter()

	// Missing required fields
	incompleteReq := `{"from": "", "to": "", "metric": ""}`

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/chart", strings.NewReader(incompleteReq))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	// The request will be parsed successfully but might cause issues in generateChart
	// This test verifies that the endpoint handles such cases gracefully
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
}

// TestGenerateChartInvalidDateFormat tests the generateChart function with invalid date format
func TestGenerateChartInvalidDateFormat(t *testing.T) {
	// Set a fixed seed for deterministic random numbers
	rand.Seed(42)

	// Create a chart request with invalid date format
	req := ChartRequest{
		From:          "01/01/2023", // Invalid format, should be YYYY-MM-DD
		To:            "31/01/2023", // Invalid format, should be YYYY-MM-DD
		Metric:        "api_to_or",
		GroupBy:       []string{"exchange_id"},
		GroupFilter:   []string{},
		CentileFilter: []string{},
	}

	// Generate chart - this should still work but might not parse dates correctly
	html, err := generateChart(req)

	// Assertions - the function should not return an error, but the chart might not have correct dates
	assert.NoError(t, err)
	assert.NotEmpty(t, html)
	assert.Contains(t, html, "api_to_or") // Title should contain metric
	// The subtitle will contain the raw date strings, not parsed dates
	assert.Contains(t, html, "From 01/01/2023 To 31/01/2023")
}

// TestGenerateChartFutureDates tests the generateChart function with future dates
func TestGenerateChartFutureDates(t *testing.T) {
	// Set a fixed seed for deterministic random numbers
	rand.Seed(42)

	// Get a date 1 year in the future
	futureYear := time.Now().Year() + 1

	// Create a chart request with future dates
	req := ChartRequest{
		From:          fmt.Sprintf("%d-01-01", futureYear),
		To:            fmt.Sprintf("%d-12-31", futureYear),
		Metric:        "api_to_or",
		GroupBy:       []string{"exchange_id"},
		GroupFilter:   []string{},
		CentileFilter: []string{},
	}

	// Generate chart
	html, err := generateChart(req)

	// Assertions
	assert.NoError(t, err)
	assert.NotEmpty(t, html)
	assert.Contains(t, html, "api_to_or") // Title should contain metric
	// The chart should be generated but might not have any data points
}

// TestGracefulShutdown tests the graceful shutdown mechanism
func TestGracefulShutdown(t *testing.T) {
	// Create a test server
	server := &http.Server{
		Addr:    ":8081", // Use a different port for testing
		Handler: setupTestRouter(),
	}

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server error: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Create a context with a timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	err := server.Shutdown(ctx)
	assert.NoError(t, err, "Server should shut down gracefully")
}
