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
	loadingDelay     = 100 * time.Millisecond
	defaultMaxWorker = 10
)

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

	// Determine the number of workers based on the number of .dem files and available CPU cores
	numFiles := len(demoFiles)
	maxWorkers := runtime.NumCPU()
	if numFiles < maxWorkers {
		maxWorkers = numFiles
	}

	fmt.Printf("Found %d .dem files. Using %d workers optimized for %d CPU cores.\n", numFiles, maxWorkers, runtime.NumCPU())

	startTime := time.Now()

	loadingDone := make(chan struct{})
	go updateLoadingAnimation(loadingDone)
	defer close(loadingDone)

	var (
		wg         sync.WaitGroup
		work       = make(chan string)
		processed  int
		totalFiles = len(demoFiles)
		corrupted  []string
	)

	printProgress(processed, totalFiles)

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for demoPath := range work {
				outputFileName := strings.TrimSuffix(filepath.Base(demoPath), filepath.Ext(demoPath)) + outputExtension
				outputPath := filepath.Join(outputFolder, outputFileName)
				errCh := make(chan error, 1)
				defer close(errCh)
				processor.ProcessDemo(demoPath, outputPath, errCh)
				if err := <-errCh; err != nil {
					corrupted = append(corrupted, demoPath)
					deleteOutputFile(outputPath)
				} else {
					processed++
				}
			}
		}()
	}

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

	wg.Wait()

	elapsed := time.Since(startTime)

	fmt.Printf("\nProcessed %d files in %s\n", totalFiles, elapsed)

	if len(corrupted) > 0 {
		fmt.Println("\nCorrupted demo files:")
		for _, file := range corrupted {
			fmt.Println(file)
		}
	}
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

func deleteOutputFile(outputPath string) {
	err := os.Remove(outputPath)
	if err != nil {
		fmt.Printf("Error deleting output file %s: %v\n", outputPath, err)
	}
}