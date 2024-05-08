package processor

import (
	"fmt"
	"time"
	"os"
	
	demoinfocs "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	common "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	events "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
)


// RegisterKillHandler registers the event handler for kill events
func RegisterKillHandler(parser demoinfocs.Parser, f *os.File) {
	parser.RegisterEventHandler(func(e events.Kill) {
		// Your code for handling kill event
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
}