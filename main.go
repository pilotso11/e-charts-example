package main

import (
	"bytes"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
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

// generateChart creates a chart using go-echarts based on the request parameters
func generateChart(req ChartRequest) (string, error) {
	// Initialize random number generator
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	// Create a new line chart
	line := charts.NewLine()

	// Set chart title and subtitle
	title := req.Metric
	subtitle := "From " + req.From + " To " + req.To

	// Set global options
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeWesteros,
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    title,
			Subtitle: subtitle,
		}),
	)

	// Enable tooltip
	line.SetGlobalOptions(charts.WithTooltipOpts(opts.Tooltip{
		Trigger: "axis",
	}))

	// Enable legend
	line.SetGlobalOptions(charts.WithLegendOpts(opts.Legend{}))

	// Generate mock data
	days := 30 // Mock data for 30 days
	xAxis := make([]string, 0)

	// Generate x-axis labels (dates)
	startDate, _ := time.Parse("2006-01-02", req.From)
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		if date.After(time.Now()) {
			break
		}
		xAxis = append(xAxis, date.Format("2006-01-02"))
	}

	// Set x-axis
	line.SetXAxis(xAxis)

	// Generate series data for each group
	for _, group := range req.GroupBy {
		// Generate random data for this group
		data := make([]opts.LineData, 0)
		for i := 0; i < len(xAxis); i++ {
			// Generate random value between 10 and 100
			value := random.Intn(90) + 10
			data = append(data, opts.LineData{Value: value})
		}

		// Add the series to the chart
		line.AddSeries(group, data)
	}

	// If no groups were selected, add a default series
	if len(req.GroupBy) == 0 {
		data := make([]opts.LineData, 0)
		for i := 0; i < len(xAxis); i++ {
			value := random.Intn(90) + 10
			data = append(data, opts.LineData{Value: value})
		}
		line.AddSeries("Default", data)
	}

	// Set line style options
	line.SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{}),
		charts.WithLabelOpts(opts.Label{}),
	)

	// Render the chart to HTML
	buffer := bytes.Buffer{}
	err := line.Render(&buffer)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
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

		// Generate a chart using go-echarts
		html, err := generateChart(chartReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ChartResponse{
				Status:  "error",
				Message: "Failed to generate chart: " + err.Error(),
				Data:    nil,
			})
			return
		}

		// Return the HTML content
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, html)
	})

	log.Println("Server starting on http://localhost:8080")
	router.Run(":8080")
}
