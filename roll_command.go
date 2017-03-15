package age_of_war

import (
	"errors"
	"fmt"

	"github.com/brdgme-go/brdgme"
)

func (g *Game) RollCommand(
	player int,
	input *brdgme.Reader,
) ([]brdgme.Log, error) {
	args, err := input.ReadLineArgs()
	if err != nil {
		return nil, fmt.Errorf("unable to check arguments to roll command: %s", err)
	}
	if len(args) > 0 {
		return nil, errors.New("didn't expect any argument for roll")
	}
	logs, err := g.RollForPlayer(player)
	return logs, err
}

func RollCommandUsage() string {
	return "{{b}}roll{{_b}} to discard a die and roll the rest"
}

func (g *Game) CanRoll(player int) bool {
	return g.CurrentPlayer == player
}

func (g *Game) RollForPlayer(player int) ([]brdgme.Log, error) {
	if !g.CanRoll(player) {
		return nil, errors.New("unable to roll right now")
	}
	logs := []brdgme.Log{g.Roll(len(g.CurrentRoll) - 1)}
	_, endOfTurnLogs := g.CheckEndOfTurn()
	logs = append(logs, endOfTurnLogs...)
	return logs, nil
}
