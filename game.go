package age_of_war

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/brdgme-go/brdgme"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

type Game struct {
	CurrentPlayer int
	Players       int

	Conquered    map[int]bool
	CastleOwners map[int]int

	CurrentlyAttacking int
	CompletedLines     map[int]bool
	CurrentRoll        []int
}

func (g *Game) Command(player int, input string, playerNames []string) ([]brdgme.Log, string, error) {
	cr := brdgme.NewReader(bytes.NewBufferString(input))
	cr.ReadSpace()
	command, err := cr.ReadWord()
	cr.ReadSpace()
	if err != nil {
		return nil, input, fmt.Errorf("unable to read command: %s", err)
	}
	var (
		logs []brdgme.Log
	)
	switch strings.ToLower(command) {
	case "attack":
		logs, err = g.AttackCommand(player, cr)
	case "line":
		logs, err = g.LineCommand(player, cr)
	case "roll":
		logs, err = g.RollCommand(player, cr)
	default:
		return logs, input, fmt.Errorf("unknown command: %s", command)
	}
	remaining := input
	if err == nil {
		if remaining, err = cr.ReadAll(); err != nil {
			return logs, input, fmt.Errorf("unable to read remaining command: %s", err)
		}
	}
	return logs, remaining, err
}

func (g *Game) Name() string {
	return "Age of War"
}

func (g *Game) Identifier() string {
	return "age_of_war"
}

func (g *Game) Encode() ([]byte, error) {
	return json.Marshal(g)
}

func (g *Game) Decode(data []byte) error {
	return json.Unmarshal(data, g)
}

func (g *Game) Start(players int) ([]brdgme.Log, error) {
	if players < 2 || players > 6 {
		return nil, errors.New("only for 2 to 6 players")
	}
	g.Players = players

	g.Conquered = map[int]bool{}
	g.CastleOwners = map[int]int{}
	g.CompletedLines = map[int]bool{}

	return []brdgme.Log{g.StartTurn()}, nil
}

func (g *Game) StartTurn() brdgme.Log {
	g.CurrentlyAttacking = -1
	g.CompletedLines = map[int]bool{}
	return g.Roll(7)
}

func (g *Game) NextTurn() brdgme.Log {
	g.CurrentPlayer = (g.CurrentPlayer + 1) % g.Players
	return g.StartTurn()
}

func (g *Game) CheckEndOfTurn() (bool, []brdgme.Log) {
	logs := []brdgme.Log{}
	if g.CurrentlyAttacking != -1 {
		c := Castles[g.CurrentlyAttacking]
		lines := c.CalcLines(
			g.Conquered[g.CurrentlyAttacking],
		)
		// If the player has completed all lines, they take the card and it is
		// the end of the turn.
		allLines := true
		for l := range lines {
			if !g.CompletedLines[l] {
				allLines = false
				break
			}
		}
		if allLines {
			suffix := ""
			if g.Conquered[g.CurrentlyAttacking] {
				suffix = fmt.Sprintf(
					" from {{player %d}}",
					g.CastleOwners[g.CurrentlyAttacking],
				)
			}
			logs = append(logs, brdgme.NewPublicLog(fmt.Sprintf(
				"{{player %d}} conquered the castle %s%s",
				g.CurrentPlayer,
				c.RenderName(),
				suffix,
			)))
			g.Conquered[g.CurrentlyAttacking] = true
			g.CastleOwners[g.CurrentlyAttacking] = g.CurrentPlayer
			if clanConquered, _ := g.ClanConquered(c.Clan); clanConquered {
				logs = append(logs, brdgme.NewPublicLog(fmt.Sprintf(
					"{{player %d}} conquered the clan %s",
					g.CurrentPlayer,
					RenderClan(c.Clan),
				)))
			}
			logs = append(logs, g.NextTurn())
			return true, logs
		}

		// If the player doesn't have enough dice to complete the rest of the
		// lines, it is the end of the turn.
		reqDice := 0
		numDice := len(g.CurrentRoll)
		canAffordLine := false
		for i, l := range lines {
			if g.CompletedLines[i] {
				continue
			}
			reqDice += l.MinDice()
			if reqDice > numDice {
				logs = append(logs, g.FailedAttackMessage(), g.NextTurn())
				return false, logs
			}
			if can, _ := l.CanAfford(g.CurrentRoll); can {
				canAffordLine = true
			}
		}

		// If the player has the minimum required dice but they can't afford a
		// line, it is the end of the turn.
		if reqDice == numDice && !canAffordLine {
			logs = append(logs, g.FailedAttackMessage(), g.NextTurn())
			return false, logs
		}
	} else {
		// If the player doesn't have enough dice for any castle, it is the end
		// of the turn.
		for i, c := range Castles {
			if g.Conquered[i] && g.CastleOwners[i] == g.CurrentPlayer {
				// They already own it.
				continue
			}
			if conquered, _ := g.ClanConquered(c.Clan); conquered {
				// The clan is conquered, can't steal.
				continue
			}
			minDice := c.MinDice()
			if g.Conquered[i] {
				minDice++
			}
			if minDice <= len(g.CurrentRoll) {
				// They can afford this one
				return false, logs
			}
		}
		// They couldn't afford anything, next turn.
		logs = append(logs, g.FailedAttackMessage(), g.NextTurn())
		return false, logs
	}
	return false, logs
}

func (g *Game) FailedAttackMessage() brdgme.Log {
	target := "anything"
	if g.CurrentlyAttacking != -1 {
		target = Castles[g.CurrentlyAttacking].RenderName()
	}
	return brdgme.NewPublicLog(fmt.Sprintf(
		"{{player %d}} failed to conquer %s",
		g.CurrentPlayer,
		target,
	))
}

func (g *Game) Scores() map[int]int {
	scores := map[int]int{}
	conqueredClans := map[int]bool{}
	for cIndex, c := range Castles {
		if !g.Conquered[cIndex] {
			continue
		}
		clanConquered, ok := conqueredClans[c.Clan]
		if !ok {
			var conqueredBy int
			clanConquered, conqueredBy = g.ClanConquered(c.Clan)
			conqueredClans[c.Clan] = clanConquered
			if clanConquered {
				scores[conqueredBy] += ClanSetPoints[c.Clan]
			}
		}
		if clanConquered {
			continue
		}
		scores[g.CastleOwners[cIndex]] += c.Points
	}
	return scores
}

func (g *Game) IsFinished() bool {
	return len(g.Conquered) == len(Castles)
}

func (g *Game) Winners() []int {
	if !g.IsFinished() {
		return []int{}
	}
	// Winner is determined by score, with ties broken by conquered clans.
	playerConqueredClans := map[int]int{}
	for _, clan := range Clans {
		if conquered, by := g.ClanConquered(clan); conquered {
			playerConqueredClans[by]++
		}
	}
	maxScore := 0
	winners := []int{}
	for p, s := range g.Scores() {
		score := s*10 + playerConqueredClans[p]
		if p == 0 || score > maxScore {
			maxScore = score
			winners = []int{}
		}
		if score == maxScore {
			winners = append(winners, p)
		}
	}
	return winners
}

func (g *Game) WhoseTurn() []int {
	return []int{g.CurrentPlayer}
}

func (g *Game) ClanConquered(clan int) (conquered bool, player int) {
	player = -1
	conquered = true
	for i, c := range Castles {
		if c.Clan != clan {
			continue
		}
		if !g.Conquered[i] {
			conquered = false
			return
		}
		if player == -1 {
			player = g.CastleOwners[i]
		} else if player != g.CastleOwners[i] {
			conquered = false
			return
		}
	}
	return
}
