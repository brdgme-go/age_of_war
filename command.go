package age_of_war

import (
	"github.com/brdgme-go/brdgme"
)

type attackCommand struct {
	castle int
}

type lineCommand struct {
	line int
}

type rollCommand struct{}

func (g *Game) CommandParser(player int) brdgme.Parser {
	oneOf := brdgme.OneOf{}
	if g.CanAttack(player) {
		oneOf = append(oneOf, attackParser)
	}
	if g.CanLine(player) {
		oneOf = append(oneOf, lineParser)
	}
	if g.CanRoll(player) {
		oneOf = append(oneOf, rollParser)
	}
	return oneOf
}

func (g *Game) CommandSpec(player int) *brdgme.Spec {
	spec := g.CommandParser(player).ToSpec()
	return &spec
}

var castleParser = brdgme.Enum{
	Values: castleEnumParserValues,
}

var attackParser = brdgme.Map{
	Parser: brdgme.Chain{
		brdgme.Token("attack"),
		brdgme.AfterSpace(castleParser),
	},
	Func: func(value interface{}) interface{} {
		return attackCommand{
			castle: value.([]interface{})[1].(int),
		}
	},
}

var lineParser = brdgme.Map{
	Parser: brdgme.Chain{
		brdgme.Token("line"),
		brdgme.AfterSpace(brdgme.Int{}),
	},
	Func: func(value interface{}) interface{} {
		return lineCommand{
			line: value.([]interface{})[1].(int),
		}
	},
}

var rollParser = brdgme.Map{
	Parser: brdgme.Token("roll"),
	Func: func(value interface{}) interface{} {
		return rollCommand{}
	},
}
