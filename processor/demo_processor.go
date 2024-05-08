package processor

import (
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
func ProcessDemo(demoPath, outputPath string) {
    // Open the file for writing demo information in CSV format
    f, err := os.Create(outputPath)
    if err != nil {
        panic(err)
    }
    defer f.Close() // Ensure file is closed even on errors

    // Write headers for CSV file
    headers := "Event,Tick,Time,Round,RoundWinner,#CT,#T,#dif,$CT,$T,$dif,Weapon,Killer,Killer.X,Killer.Y,Victim,Victim.X,Victim.Y,CT#1 X,CT#1 Y,CT#2 X,CT#2 Y,CT#3 X,CT#3 Y,CT#4 X,CT#4 Y,CT#5 X,CT#5 Y,T#1 X,T#1 Y,T#2 X,T#2 Y,T#3 X,T#3 Y,T#4 X,T#4 Y,T#5 X,T#5 Y\n"
    _, err = f.WriteString(headers)
    if err != nil {
        panic(err)
    }

    // Open the demo file
    demoFile, err := os.Open(demoPath)
    if err != nil {
        panic(err)
    }
    defer demoFile.Close() // Close the demo file after processing

    // Create a new parser for the demo file
    parser := demoinfocs.NewParser(demoFile)
    defer parser.Close() // Close the parser after processing

    // Register event handlers for processing demo events
    RegisterRoundStartHandler(parser)
    RegisterKillHandler(parser, f) // Pass the file object to RegisterKillHandler
    RegisterRoundEndHandler(parser, f) // Pass the file object to RegisterRoundEndHandler

    // Parse the demo file to process all events
    err = parser.ParseToEnd()
    checkError(err)
}

func checkError(err error) {
    if err != nil {
        panic(err)
    }
}