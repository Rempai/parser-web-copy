package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	demoinfocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	common "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	events "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
)

const (
	demoFolder       = "demo-in-extra"
	outputFolder     = "demo-out"
	outputExtension  = ".csv"
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
				processDemo(demoPath, outputPath)
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
	// loading := getLoadingAnimation()
	// fmt.Printf("\r%s Processed %d/%d demo files", loading, processed, total)
}

func updateLoadingAnimation(done chan struct{}) {
	animationIndex := 0
	ticker := time.NewTicker(loadingDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// fmt.Printf("\r%s", loadingPatterns[animationIndex])
			animationIndex = (animationIndex + 1) % len(loadingPatterns)
		case <-done:
			return
		}
	}
}

func getLoadingAnimation() string {
	return loadingPatterns[0]
}

func formatPlayer(p *common.Player) string {
	if p == nil {
		return "?"
	}
	switch p.Team {
	case common.TeamTerrorists:
		return "[T]" + p.Name
	case common.TeamCounterTerrorists:
		return "[CT]" + p.Name
	}
	return p.Name
}

func processDemo(demoPath, outputPath string) {
	// Open the file for writing demo information in CSV format
	f, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer f.Close() // Ensure file is closed even on errors

	// Write headers for CSV file including the new column for round number and win information
	// headers := "Event,Round,Tick,CTPlayers,TPlayers,TeamDifference,Killer,Victim,Weapon,Winner,CTMoney,TMoney,MoneyDifference,KillerX,KillerY,VictimX,VictimY,KillDistance\n"
	headers := "Event,Tick,Time,Round,RoundWinner,#CT,#T,#dif,$CT,$T,$dif,Weapon,Killer,Killer.X,Killer.Y,Victim,Victim.X,Victim.Y,CT#1 X,CT#1 Y,CT#2 X,CT#2 Y,CT#3 X,CT#3 Y,CT#4 X,CT#4 Y,CT#5 X,CT#5 Y,T#1 X,T#1 Y,T#2 X,T#2 Y,T#3 X,T#3 Y,T#4 X,T#4 Y,T#5 X,T#5 Y\n"
	_, err = f.WriteString(headers)
	if err != nil {
		panic(err)
	}

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
	// fmt.Println(myvars)

	// Open the demo file
	demoFile, err := os.Open(demoPath)
	if err != nil {
		panic(err)
	}
	defer demoFile.Close() // Close the demo file after processing

	// Create a new parser for the demo file
	parser := demoinfocs.NewParser(demoFile)
	defer parser.Close() // Close the parser after processing

	// Define helper function to handle errors
	checkError := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	var (
	// ctPlayers, tPlayers int
	// roundNumber         = 0
	// roundWinners        = make(map[int]string) // Map to store round winners by round number
	)

	// Register event handlers for processing demo events
	parser.RegisterEventHandler(func(e events.RoundStart) {
		csv.Round++
		// roundNumber++ // Increment round number at the start of each round
		csv.CTs = 5
		// ctPlayers = 5
		csv.Ts = 5
		// tPlayers = 5
	})

	parser.RegisterEventHandler(func(e events.Kill) {
		csv.Event = "Kill"
		csv.Tick = parser.GameState().IngameTick()
		csv.Weapon = e.Weapon.String()

		// Update player counts based on team of victim
		if e.Victim.Team == common.TeamCounterTerrorists {
			csv.CTs--
		} else if e.Victim.Team == common.TeamTerrorists {
			csv.Ts--
		}
		// Calculate the difference between the number of players on each team.
		csv.PlayerDif = csv.CTs - csv.Ts

		csv.CTEqVal = parser.GameState().TeamCounterTerrorists().CurrentEquipmentValue()
		csv.TEqVal = parser.GameState().TeamTerrorists().CurrentEquipmentValue()
		csv.EqValDif = csv.CTEqVal - csv.TEqVal

		var timer = parser.CurrentTime().Round(6 * time.Second)
		csv.Time = timer.String()
		// fmt.Println(myvars.Time)

		// Track player names and positions, with an extra entry for killers and victims.
		type PlayerPosition struct {
			Name string
			X    int
			Y    int
		}
		type KillEventPositions struct {
			Killer PlayerPosition
			Victim PlayerPosition
		}

		var killEventPositions KillEventPositions
		var CTPositions [5]PlayerPosition
		var TPositions [5]PlayerPosition

		// There is no Killer for C4 kills.
		if e.Weapon.String() != "C4" {
			killEventPositions.Killer = PlayerPosition{e.Killer.Name, int(e.Killer.Position().X), int(e.Killer.Position().Y)}
		}
		killEventPositions.Victim = PlayerPosition{e.Victim.Name, int(e.Victim.LastAlivePosition.X), int(e.Victim.LastAlivePosition.Y)}

		// Loop through both Teams, get names and locations.
		for i, member := range parser.GameState().TeamCounterTerrorists().Members() {
			if member.IsAlive() {
				CTPositions[i].Name = member.Name
				CTPositions[i].X, CTPositions[i].Y = int(member.Position().X), int(member.Position().Y)
			}
		}

		// Loop through Counter-Terrorists and extract names
		for i, member := range parser.GameState().TeamTerrorists().Members() {
			if member.IsAlive() {
				TPositions[i].Name = member.Name
				TPositions[i].X, TPositions[i].Y = int(member.Position().X), int(member.Position().Y)
			}
		}

		// Write the event information to the file including the round number, team difference, and other details
		// fmt.Println(csv)
		csv.Killer = killEventPositions.Killer.Name
		csv.KillerX = killEventPositions.Killer.X
		csv.KillerY = killEventPositions.Killer.Y
		csv.Victim = killEventPositions.Victim.Name
		csv.VictimX = killEventPositions.Victim.X
		csv.VictimY = killEventPositions.Victim.Y

		// headers := "Event,Tick,Time,Round,RoundWinner,#CT,#T,#dif,$CT,$T,$dif,Weapon,Killer,Killer.X,Killer.Y,Victim,Victim.X,Victim.Y,CT[1].X,CT[1].Y,CT[2].X,CT[2].Y,CT[3].X,CT[3].Y,CT[4].X,CT[4].Y,CT[5].X,CT[5].Y,CT[1].X,CT[1].Y,CT[2].X,CT[2].Y,CT[3].X,CT[3].Y,CT[4].X,CT[4].Y,CT[5].X,CT[5].Y\n"
		// fmt.Println(stringtowrite)
		csvLine := fmt.Sprintf("%s,%d,%s,%d,%d,%d,%d,%d,%d,%d,%d,%s,%s,%d,%d,%s,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d\n", csv.Event, csv.Tick, csv.Time, csv.Round, csv.RoundWinner, csv.CTs, csv.Ts, csv.PlayerDif, csv.CTEqVal, csv.TEqVal, csv.EqValDif, csv.Weapon, csv.Killer, csv.KillerX, csv.KillerY, csv.Victim, csv.VictimX, csv.VictimY, CTPositions[0].X, CTPositions[0].Y, CTPositions[1].X, CTPositions[1].Y, CTPositions[2].X, CTPositions[2].Y, CTPositions[3].X, CTPositions[3].Y, CTPositions[4].X, CTPositions[4].Y, TPositions[0].X, TPositions[0].Y, TPositions[1].X, TPositions[1].Y, TPositions[2].X, TPositions[2].Y, TPositions[3].X, TPositions[3].Y, TPositions[4].X, TPositions[4].Y)
		fmt.Println(csvLine)
		_, err := f.WriteString(csvLine)

		// msg := fmt.Sprintf("Kill,%d,%d,%d,%d,%d,%s,%s,%s,%s,%d,%d,%s,%d,%d,%d,%d\n",
		// 	roundNumber, parser.GameState().IngameTick(), ctPlayers, tPlayers, csv.PlayerDif, formatPlayer(e.Killer), formatPlayer(e.Victim), e.Weapon.String(), "", csv.CTEqVal, csv.TEqVal, csv.Time, killEventPositions.Killer.X, killEventPositions.Killer.Y, killEventPositions.Victim.X, killEventPositions.Victim.Y)
		// _, err := f.WriteString(msg)
		checkError(err)
	})

	parser.RegisterEventHandler(func(e events.RoundEnd) {
		// Determine the winner of the round
		switch e.Winner {
		case common.TeamTerrorists:
			csv.RoundWinner = -1
			// roundWinners[roundNumber] = "-1" // Ts win, so the winner value is negative
		case common.TeamCounterTerrorists:
			csv.RoundWinner = 1
			// roundWinners[roundNumber] = "1" // CTs win, so the winner value is positive
		}

		// Write round end event to the file with the current round winner
		// msg := fmt.Sprintf("RoundEnd,%d,%d,%d,%d,%s,,,,,%d,%d,%d\n", roundNumber, parser.GameState().IngameTick(), ctPlayers, tPlayers, roundWinners[roundNumber], ctMoney, tMoney, ctMoney-tMoney)
		// _, err := f.WriteString(msg)
		skipline := "\n"
		f.WriteString(skipline)
		checkError(err)
	})

	// parser.RegisterEventHandler(func(e events.BombPlanted) {
	// 	msg := fmt.Sprintf("BombPlanted,%d,%d,%d,%d,%d,,,,%s,%d,%d,%d\n", roundNumber, parser.GameState().IngameTick(), ctPlayers, tPlayers, ctPlayers-tPlayers, "Planted", ctMoney, tMoney, ctMoney-tMoney)
	// 	_, err := f.WriteString(msg) // Write bomb planted information to file in CSV format
	// 	checkError(err)
	// })

	// parser.RegisterEventHandler(func(e events.BombDefused) {
	// 	msg := fmt.Sprintf("BombDefused,%d,%d,%d,%d,%d,,,,%s,%d,%d,%d\n", roundNumber, parser.GameState().IngameTick(), ctPlayers, tPlayers, ctPlayers-tPlayers, "Defused", ctMoney, tMoney, ctMoney-tMoney)
	// 	_, err := f.WriteString(msg) // Write bomb defused information to file in CSV format
	// 	checkError(err)
	// })

	// parser.RegisterEventHandler(func(e events.BombExplode) {
	// 	msg := fmt.Sprintf("BombExplode,%d,%d,%d,%d,%d,,,,%s,%d,%d,%d\n", roundNumber, parser.GameState().IngameTick(), ctPlayers, tPlayers, ctPlayers-tPlayers, "Exploded", ctMoney, tMoney, ctMoney-tMoney)
	// 	_, err := f.WriteString(msg) // Write bomb exploded information to file in CSV format
	// 	checkError(err)
	// })

	// Parse the demo file to process all events
	err = parser.ParseToEnd()
	checkError(err)
}
