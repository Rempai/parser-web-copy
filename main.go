package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"parser-systeem/processing"
)

const (
	demoFolder   = "demo-in"
	outputFolder = "csv-out"
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
	go processing.UpdateLoadingAnimation(loadingDone) // Adjusted to exported function name
	defer close(loadingDone)

	processing.StartWorkers(maxWorkers, demoFiles, demoFolder, outputFolder) // Adjusted to exported function name

	elapsed := time.Since(startTime)

	fmt.Printf("\nProcessed %d files in %s\n", len(demoFiles), elapsed)
}
