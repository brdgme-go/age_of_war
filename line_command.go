package age_of_war

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/brdgme-go/brdgme"
)

func (g *Game) LineCommand(
	player int,
	input *brdgme.Reader,
) ([]brdgme.Log, error) {
	args, err := input.ReadLineArgs()
	if err != nil || len(args) != 1 {
		return nil, errors.New("please specify a line to complete")
	}

	line, err := strconv.Atoi(args[0])
	if err != nil || line <= 0 {
		return nil, errors.New("the line must be a number greater than 0")
	}
	logs, err := g.Line(player, line-1)
	return logs, err
}

func LineCommandUsage(player string, context interface{}) string {
	return "{{b}}line #{{_b}} to complete a line in the castle you are attacking, eg. {{b}}line 2{{_b}}"
}

func (g *Game) CanLine(player int) bool {
	return g.CurrentPlayer == player && g.CurrentlyAttacking != -1
}

func (g *Game) Line(player, line int) ([]brdgme.Log, error) {
	if !g.CanLine(player) {
		return nil, errors.New("unable to complete a line right now")
	}
	lines := Castles[g.CurrentlyAttacking].CalcLines(
		g.Conquered[g.CurrentlyAttacking],
	)
	if line < 0 || line >= len(lines) {
		return nil, errors.New("that is not a valid line")
	}
	if g.CompletedLines[line] {
		return nil, errors.New("that line has already been completed")
	}
	canAfford, with := lines[line].CanAfford(g.CurrentRoll)
	if !canAfford {
		return nil, errors.New("cannot afford that line")
	}
	logs := []brdgme.Log{
		brdgme.NewPublicLog(fmt.Sprintf(
			"{{player %d}} completed %s with {{b}}%d{{_b}} %s",
			player,
			lines[line].String(),
			with,
			brdgme.Plural(with, "die"),
		)),
	}
	g.CompletedLines[line] = true
	// Check end of turn first in case they completed the castle.
	isEndOfTurn, endOfTurnLogs := g.CheckEndOfTurn()
	logs = append(logs, endOfTurnLogs...)
	if !isEndOfTurn {
		logs = append(logs, g.Roll(len(g.CurrentRoll)-with))
		_, endOfTurnLogs = g.CheckEndOfTurn()
		logs = append(logs, endOfTurnLogs...)
	}
	return logs, nil
}
