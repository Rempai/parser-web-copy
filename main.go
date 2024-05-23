package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"parser-systeem/processor"
)

const (
	demoFolder   = "demo-in"
	outputFolder = "csv-out"
	staticFolder = "generated" // SvelteKit build output folder
)

func main() {
	fmt.Println("Started processing demo files...")

	// Create the necessary folders if they don't exist
	if err := os.MkdirAll(demoFolder, os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(outputFolder, os.ModePerm); err != nil {
		panic(err)
	}

	// Set up the Gin server
	r := gin.Default()

	// Enable CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"} // Update with your Svelte frontend URL
	r.Use(cors.New(config))

	// API endpoint to upload files
	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
			return
		}

		// Save the uploaded file
		filePath := filepath.Join(demoFolder, file.Filename)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save the file"})
			return
		}

		// Process the uploaded file
		outputFile, err := processFile(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Serve the processed CSV file
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(outputFile)))
		c.File(outputFile)
	})

	// Serve static files from the SvelteKit build directory
	r.Static("/static", staticFolder)
	r.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join(staticFolder, "index.html"))
	})

	// Start the server on port 8080
	r.Run(":8080")
}

func processFile(filePath string) (string, error) {
	err := cleanupFiles()
	if err != nil {
		return "", err
	}

	// Start processing the uploaded file
	startTime := time.Now()

	// Process the demo file
	outputFilePath := filepath.Join(outputFolder, filepath.Base(filePath)+".csv")
	errCh := make(chan error, 1)
	go processor.ProcessDemo(filePath, outputFilePath, errCh)
	if err := <-errCh; err != nil {
		return "", err
	}

	elapsed := time.Since(startTime)
	fmt.Printf("\nProcessed file %s in %s\n", filepath.Base(filePath), elapsed)

	return outputFilePath, nil
}

func cleanupFiles() error {
	// Remove all files in demoFolder
	err := os.RemoveAll(demoFolder)
	if err != nil {
		return err
	}
	err = os.MkdirAll(demoFolder, os.ModePerm)
	if err != nil {
		return err
	}

	// Remove all files in outputFolder
	err = os.RemoveAll(outputFolder)
	if err != nil {
		return err
	}
	err = os.MkdirAll(outputFolder, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
