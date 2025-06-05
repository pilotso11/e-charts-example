package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// ChartRequest represents the form data sent to the /chart endpoint
type ChartRequest struct {
	From          string   `json:"from"`
	To            string   `json:"to"`
	Metric        string   `json:"metric"`
	GroupBy       []string `json:"groupBy"`
	GroupFilter   []string `json:"groupFilter"`
	CentileFilter []string `json:"centileFilter"`
}

// ChartResponse represents the response from the /chart endpoint
type ChartResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func main() {
	// Create a default gin router
	router := gin.Default()

	// Ensure static directory exists
	if err := os.MkdirAll("static", 0755); err != nil {
		log.Fatalf("Failed to create static directory: %v", err)
	}

	// Serve static files
	router.Static("/", "./static")

	// Handle POST requests to /chart endpoint
	router.POST("/chart", func(c *gin.Context) {
		var chartReq ChartRequest
		if err := c.ShouldBindJSON(&chartReq); err != nil {
			c.JSON(http.StatusBadRequest, ChartResponse{
				Status:  "error",
				Message: "Invalid request body",
				Data:    nil,
			})
			return
		}

		// Return the submitted JSON directly
		c.JSON(http.StatusOK, chartReq)
	})

	log.Println("Server starting on http://localhost:8080")
	router.Run(":8080")
}
