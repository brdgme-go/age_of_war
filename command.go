package age_of_war

import "github.com/brdgme-go/brdgme"

type attackCommand struct {
	castle int
}

func (g *Game) CommandSpec(player int) *brdgme.Spec {
	oneOf := brdgme.OneOf{}
	if g.CanAttack(player) {
		oneOf = append(oneOf, g.attackParser(player))
	}
	spec := oneOf.ToSpec()
	return &spec
}

func (g *Game) attackParser(player int) brdgme.Spec {
	return brdgme.Chain{
		brdgme.Token("attack").ToSpec(),
		brdgme.AfterSpace(castleParser).ToSpec(),
	}.ToSpec()
}

var castleParser = brdgme.Enum{
	Values: castleEnumParserValues,
}.ToSpec()

var attackParser = brdgme.Map{
	Spec: brdgme.Chain{
		brdgme.Token("attack").ToSpec(),
		brdgme.AfterSpace(castleParser).ToSpec(),
	}.ToSpec(),
	Func: func(value interface{}) interface{} {
		parts := value.([]interface{})
		castle := parts[1].(int)
		return attackCommand{
			castle: castle,
		}
	},
}
