package processor

import (
	"fmt"
	"os"

	demoinfocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
)

// EventVars holds the variables related to each event
type EventVars struct {
	Event       string
	Tick        int
	Time        string
	Round       int
	RoundWinner int
	CTs         int
	Ts          int
	PlayerDif   int
	CTEqVal     int
	TEqVal      int
	EqValDif    int
	Weapon      string
	Killer      string
	KillerX     int
	KillerY     int
	Victim      string
	VictimX     int
	VictimY     int
}

var csv EventVars

// ProcessDemo processes a single demo file
func ProcessDemo(demoPath, outputPath string, errCh chan<- error) {
	fmt.Printf("Starting to process demo file: %s\n", demoPath)

	// Open the file for writing demo information in CSV format
	f, err := os.Create(outputPath)
	if err != nil {
		errCh <- fmt.Errorf("failed to create output file: %v", err)
		return
	}
	defer f.Close() // Ensure file is closed even on errors

	fmt.Printf("Created output file: %s\n", outputPath)

	// Write headers for CSV file
	headers := "Event,Tick,Time,Round,RoundWinner,#CT,#T,#dif,$CT,$T,$dif,Weapon,Killer,Killer.X,Killer.Y,Victim,Victim.X,Victim.Y,CT#1 X,CT#1 Y,CT#2 X,CT#2 Y,CT#3 X,CT#3 Y,CT#4 X,CT#4 Y,CT#5 X,CT#5 Y,T#1 X,T#1 Y,T#2 X,T#2 Y,T#3 X,T#3 Y,T#4 X,T#4 Y,T#5 X,T#5 Y\n"
	_, err = f.WriteString(headers)
	if err != nil {
		errCh <- fmt.Errorf("failed to write headers to output file: %v", err)
		return
	}

	// Open the demo file
	demoFile, err := os.Open(demoPath)
	if err != nil {
		errCh <- fmt.Errorf("failed to open demo file: %v", err)
		return
	}
	defer demoFile.Close() // Close the demo file after processing

	fmt.Printf("Opened demo file: %s\n", demoPath)

	// Create a new parser for the demo file
	parser := demoinfocs.NewParser(demoFile)
	defer parser.Close() // Close the parser after processing

	fmt.Println("Created parser for the demo file")

	// Register event handlers for processing demo events
	RegisterRoundStartHandler(parser)
	RegisterKillHandler(parser, f)     // Pass the file object to RegisterKillHandler
	RegisterRoundEndHandler(parser, f) // Pass the file object to RegisterRoundEndHandler

	// Parse the demo file to process all events
	err = parser.ParseToEnd()
	if err != nil {
		errCh <- fmt.Errorf("failed to parse demo file: %v", err)
		return
	}

	fmt.Printf("Successfully processed demo file: %s\n", demoPath)

	// No errors, signal completion
	errCh <- nil
}
