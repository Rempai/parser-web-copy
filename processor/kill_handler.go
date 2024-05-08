package processor

import (
	"fmt"
	"time"
	"os"
	
	demoinfocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	common "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	events "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
)


func RegisterKillHandler(parser demoinfocs.Parser, f *os.File) {
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

		// Write the event information to the file including the round number, team difference, and other details
		csv.Killer = killEventPositions.Killer.Name
		csv.KillerX = killEventPositions.Killer.X
		csv.KillerY = killEventPositions.Killer.Y
		csv.Victim = killEventPositions.Victim.Name
		csv.VictimX = killEventPositions.Victim.X
		csv.VictimY = killEventPositions.Victim.Y

		// Write CSV line to file
		csvLine := fmt.Sprintf("%s,%d,%s,%d,%d,%d,%d,%d,%d,%d,%d,%s,%s,%d,%d,%s,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d\n", csv.Event, csv.Tick, csv.Time, csv.Round, csv.RoundWinner, csv.CTs, csv.Ts, csv.PlayerDif, csv.CTEqVal, csv.TEqVal, csv.EqValDif, csv.Weapon, csv.Killer, csv.KillerX, csv.KillerY, csv.Victim, csv.VictimX, csv.VictimY, CTPositions[0].X, CTPositions[0].Y, CTPositions[1].X, CTPositions[1].Y, CTPositions[2].X, CTPositions[2].Y, CTPositions[3].X, CTPositions[3].Y, CTPositions[4].X, CTPositions[4].Y, TPositions[0].X, TPositions[0].Y, TPositions[1].X, TPositions[1].Y, TPositions[2].X, TPositions[2].Y, TPositions[3].X, TPositions[3].Y, TPositions[4].X, TPositions[4].Y)
		_, err := f.WriteString(csvLine)
		if err != nil {
			panic(err)
		}
	})
}