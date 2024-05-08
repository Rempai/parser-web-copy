package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"parser-systeem/processor"
)

const (
	demoFolder       = "demo-in"
	outputFolder     = "csv-out"
	outputExtension  = ".csv"
	mapFilter        = false // bool whether you wanna filter demos for a specific map
	mapToFilter      = ""    // string value of the map to filter. Options: "de_mirage",
	loadingDelay     = 100 * time.Millisecond
	defaultMaxWorker = 10
	defaultWorkers   = 2
)

// "Resources" for the console output loading animation to use
var loadingPatterns = []string{"[.  ]", "[.. ]", "[...]", "[ ..]", "[  .]"}

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

	// Calculate total file size of .dem files
	var totalFileSize int64
	for _, fileInfo := range demoFiles {
		if fileInfo.IsDir() || !strings.HasSuffix(strings.ToLower(fileInfo.Name()), ".dem") {
			continue
		}
		filePath := filepath.Join(demoFolder, fileInfo.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			panic(err)
		}
		totalFileSize += info.Size()
	}

	// Determine the number of workers based on the number of .dem files and available CPU cores
	numFiles := len(demoFiles)
	maxWorkers := runtime.NumCPU()
	if numFiles < maxWorkers {
		maxWorkers = numFiles
	}

	// Output information about worker selection
	if numFiles <= runtime.NumCPU() {
		fmt.Printf("Found %d .dem files. Using %d workers based on file count.\n", numFiles, maxWorkers)
	} else {
		fmt.Printf("Found %d .dem files. Using %d workers optimized for %d CPU cores.\n", numFiles, maxWorkers, runtime.NumCPU())
	}

	// Start time to measure total processing time
	startTime := time.Now()

	// Start goroutine to update the loading animation
	loadingDone := make(chan struct{})
	go updateLoadingAnimation(loadingDone)
	defer close(loadingDone)

	var (
		wg         sync.WaitGroup
		work       = make(chan string)
		processed  int
		totalFiles = len(demoFiles)
	)

	// Print initial progress
	printProgress(processed, totalFiles)

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for demoPath := range work {
				outputPath := filepath.Join(outputFolder, strings.TrimSuffix(filepath.Base(demoPath), filepath.Ext(demoPath))+outputExtension)
				processor.ProcessDemo(demoPath, outputPath)
				processed++
				printProgress(processed, totalFiles)
			}
		}()
	}

	// Enqueue work
	for _, fileInfo := range demoFiles {
		if fileInfo.IsDir() {
			continue // Skip directories
		}
		if strings.HasSuffix(strings.ToLower(fileInfo.Name()), ".dem") {
			demoPath := filepath.Join(demoFolder, fileInfo.Name())
			work <- demoPath
		}
	}
	close(work)

	// Wait for all workers to finish
	wg.Wait()

	// Calculate total processing time
	elapsed := time.Since(startTime)

	// Print completion message and total processing time
	// Print completion message with compact statistical presentation
	fmt.Printf("\nProcessed %d files, totaling %.1f GB in %s\n", totalFiles, float64(totalFileSize)/(1024*1024*1024), elapsed)
}

func printProgress(processed, total int) {
	loading := getLoadingAnimation()
	fmt.Printf("\r%s Processed %d/%d demo files", loading, processed, total)
}

func updateLoadingAnimation(done chan struct{}) {
	animationIndex := 0
	ticker := time.NewTicker(loadingDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Printf("\r%s", loadingPatterns[animationIndex])
			animationIndex = (animationIndex + 1) % len(loadingPatterns)
		case <-done:
			return
		}
	}
}

func getLoadingAnimation() string {
	return loadingPatterns[0]
}