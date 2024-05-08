package processor

import (
	"os"

	demoinfocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	events "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	common "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
)

// RegisterRoundEndHandler registers the event handler for round end events
func RegisterRoundEndHandler(parser demoinfocs.Parser, f *os.File) {
	parser.RegisterEventHandler(func(e events.RoundEnd) {
		// Your code for handling round end event
		switch e.Winner {
		case common.TeamTerrorists:
			csv.RoundWinner = -1
		case common.TeamCounterTerrorists:
			csv.RoundWinner = 1
		}

		// Write CSV line to file
		skipline := "\n"
		_, err := f.WriteString(skipline)
		if err != nil {
			panic(err)
		}
	})
}