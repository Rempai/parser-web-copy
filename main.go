package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"parser-web/processor"

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
	config := cors.Config{
		AllowOrigins:     []string{"https://strategy-master-tool-thuas-id-2324-spring-counte-34ebbbb5dd61de.gitlab.io"}, // Update with your Svelte frontend URL without trailing slash
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(config))

	// API endpoint to upload files
	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			fmt.Printf("Error receiving file: %v\n", err) // Log the error
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
			return
		}

		fmt.Printf("Received file: %s\n", file.Filename)

		// Obtain the original file name
		originalFileName := file.Filename

		// Save the uploaded file
		filePath := filepath.Join(demoFolder, originalFileName)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			fmt.Printf("Error saving uploaded file: %v\n", err) // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save the file"})
			return
		}

		fmt.Printf("Saved uploaded file to: %s\n", filePath)

		// Process the uploaded file
		outputFile, err := processFile(filePath)
		if err != nil {
			fmt.Printf("Error processing file: %v\n", err) // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		fmt.Printf("Processed file and generated output: %s\n", outputFile)

		// Set the Content-Type header to text/csv
		c.Header("Content-Type", "text/csv")
		// Set the Content-Disposition header for attachment with the original filename and a .csv extension
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", originalFileName[:len(originalFileName)-4]))

		// Serve the processed CSV file
		c.File(outputFile)

		// Delete the original .dem file after serving the CSV file
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Error deleting file %s: %v\n", filePath, err) // Log the error if unable to delete the file
		} else {
			fmt.Printf("Deleted file: %s\n", filePath)
		}
	})

	// Serve static files from the SvelteKit build directory
	r.Static("/static", staticFolder)
	r.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join(staticFolder, "index.html"))
	})

	// Start the server on port 8080
	r.Run(":80")
}

func processFile(filePath string) (string, error) {
	fmt.Printf("Processing file: %s\n", filePath)

	// Start processing the uploaded file
	startTime := time.Now()

	// Process the demo file
	originalFileName := filepath.Base(filePath)
	outputFileName := originalFileName[:len(originalFileName)-4] + ".csv" // Remove ".dem" extension and add ".csv"
	outputFilePath := filepath.Join(outputFolder, outputFileName)
	errCh := make(chan error, 1)
	go processor.ProcessDemo(filePath, outputFilePath, errCh)
	if err := <-errCh; err != nil {
		fmt.Printf("Error processing demo file: %v\n", err) // Log the error
		return "", err
	}

	elapsed := time.Since(startTime)
	fmt.Printf("Processed file %s in %s\n", originalFileName, elapsed)

	return outputFilePath, nil
}