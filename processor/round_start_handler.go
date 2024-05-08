package processor

import (
	demoinfocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	events "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
)

// RegisterRoundStartHandler registers the event handler for round start events
func RegisterRoundStartHandler(parser demoinfocs.Parser) {
	parser.RegisterEventHandler(func(e events.RoundStart) {
		csv.Round++
		// roundNumber++ // Increment round number at the start of each round
		csv.CTs = 5
		// ctPlayers = 5
		csv.Ts = 5
		// tPlayers = 5
	})
}