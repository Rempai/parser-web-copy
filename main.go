package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"parser-systeem/processing"
)

const (
	demoFolder   = "demo-in"
	outputFolder = "csv-out"
	staticFolder = "public" // Svelte build output folder
)

func main() {
	fmt.Println("Started processing demo files...")

	// Create the output folder if it doesn't exist
	if err := os.MkdirAll(outputFolder, os.ModePerm); err != nil {
		panic(err)
	}

	// Get all demo files in the demo folder
	demoFiles, err := os.ReadDir(demoFolder)
	if err != nil {
		panic(err)
	}

	// Determine the number of workers based on the number of .dem files and available CPU cores
	numFiles := len(demoFiles)
	maxWorkers := runtime.NumCPU()
	if numFiles < maxWorkers {
		maxWorkers = numFiles
	}

	fmt.Printf("Found %d .dem files. Using %d workers optimized for %d CPU cores.\n", numFiles, maxWorkers, runtime.NumCPU())

	startTime := time.Now()

	loadingDone := make(chan struct{})
	go processing.UpdateLoadingAnimation(loadingDone)
	defer close(loadingDone)

	processing.StartWorkers(maxWorkers, demoFiles, demoFolder, outputFolder)

	elapsed := time.Since(startTime)
	fmt.Printf("\nProcessed %d files in %s\n", len(demoFiles), elapsed)

	// Set up the Gin server
	r := gin.Default()

	// API endpoint
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello from Go",
		})
	})

	// Serve static files from the Svelte build directory
	r.Static("/static", staticFolder)
	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(staticFolder, "index.html"))
	})

	// Start the server on port 8080
	r.Run(":8080")
}