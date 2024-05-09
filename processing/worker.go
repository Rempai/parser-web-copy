package processing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"parser-systeem/processor"
)

func StartWorkers(maxWorkers int, demoFiles []os.DirEntry, demoFolder, outputFolder string) {
	var (
		wg         sync.WaitGroup
		work       = make(chan string)
		processed  int
		totalFiles = len(demoFiles)
		corrupted  []string
	)

	PrintProgress(processed, totalFiles)

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh := make(chan error, 1) // Move channel creation here
			defer close(errCh)           // Close the channel when the goroutine exits
			for demoPath := range work {
				outputFileName := strings.TrimSuffix(filepath.Base(demoPath), filepath.Ext(demoPath)) + outputExtension
				outputPath := filepath.Join(outputFolder, outputFileName)
				processor.ProcessDemo(demoPath, outputPath, errCh)
				if err := <-errCh; err != nil {
					corrupted = append(corrupted, demoPath)
					DeleteOutputFile(outputPath)
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

	fmt.Printf("\nProcessed %d files\n", totalFiles)

	if len(corrupted) > 0 {
		fmt.Println("\nCorrupted demo files:")
		for _, file := range corrupted {
			fmt.Println(file)
		}
	}
}
