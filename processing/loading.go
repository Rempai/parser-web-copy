package processing

import (
	"fmt"
	"os"
	"time"
)

const (
	outputExtension = ".csv"
	loadingDelay    = 100 * time.Millisecond
)

var loadingPatterns = []string{"[.  ]", "[.. ]", "[...]", "[ ..]", "[  .]"}

func PrintProgress(processed, total int) {
	loading := GetLoadingAnimation()
	fmt.Printf("\r%s Processed %d/%d demo files", loading, processed, total)
}

func UpdateLoadingAnimation(done chan struct{}) {
	animationIndex := 0
	ticker := time.NewTicker(loadingDelay) // Use loadingDelay here
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

func GetLoadingAnimation() string {
	return loadingPatterns[0]
}

func DeleteOutputFile(outputPath string) {
	err := os.Remove(outputPath)
	if err != nil {
		fmt.Printf("Error deleting output file %s: %v\n", outputPath, err)
	}
}
