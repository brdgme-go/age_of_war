package age_of_war

import (
	"errors"
	"fmt"

	"github.com/brdgme-go/brdgme"
	"github.com/brdgme-go/render"
)

func (g *Game) AttackCommand(
	pNum int,
	input *brdgme.Reader,
) ([]brdgme.Log, bool, error) {
	args, err := input.ReadLineArgs()
	if err != nil || len(args) != 1 {
		return nil, false, errors.New("please specify a castle to attack")
	}

	castleNames := []string{}
	for _, c := range Castles {
		castleNames = append(castleNames, c.Name)
	}
	castle, err := brdgme.MatchStringInStrings(args[0], castleNames)
	if err != nil {
		return nil, false, err
	}

	logs, err := g.Attack(pNum, castle)
	return logs, true, err
}

func AttackUsage() string {
	return "{{b}}attack #{{_b}} to attack a castle, eg. {{b}}attack kita{{_b}}"
}

func (g *Game) CanAttack(player int) bool {
	return g.CurrentPlayer == player && g.CurrentlyAttacking == -1
}

func (g *Game) Attack(player, castle int) ([]brdgme.Log, error) {
	if !g.CanAttack(player) {
		return nil, errors.New("unable to attack a castle right now")
	}
	if castle < 0 || castle >= len(Castles) {
		return nil, errors.New("that is not a valid castle")
	}
	if g.Conquered[castle] && g.CastleOwners[castle] == player {
		return nil, errors.New("you have already conquered that castle")
	}
	if ok, _ := g.ClanConquered(Castles[castle].Clan); ok {
		return nil, errors.New("that clan is already conquered")
	}
	g.CurrentlyAttacking = castle
	logs := []brdgme.Log{brdgme.NewPublicLog(fmt.Sprintf(
		"%s is attacking:\n%s",
		render.Player(player),
		g.RenderCastle(castle, []int{}),
	))}
	_, endLogs := g.CheckEndOfTurn()
	logs = append(logs, endLogs...)
	return logs, nil
}
