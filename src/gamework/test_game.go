/*

TestGameEngine implements an extremely simple "game".  On each player's
turn, he can act or pass.  If he acts, he can act again, and keep acting
until he passes.  Finally, a player can quit, which ends the game.

The purpose of this is to facilitate writing tests of the GameEngine,
particularly the serialization and deserialization features.

*/
package gamework

import (
	"fmt"
)

type TestGameEngine struct {
	id           string
	name         string
	players      []Player
	actingPlayer int
	actionCount  int
}

func InitTestGameWithTestEngine() Game {
	players := make([]Player, 2)
	players[0] = Player{Id: "A", Name: "Allen"}
	players[1] = Player{Id: "B", Name: "Bob"}
	g := Game{
		Id:      "T",
		Name:    "Test",
		Players: players,
		Engine:  new(TestGameEngine)}

	return g
}

func (e0 *TestGameEngine) Equals(e GameEngine) bool {
	// Since there's no such thing as a pointer to an interface, how do you assert
	// that e (which is an interface) is actually a *TestGameEngine (a pointer to
	// a type that implements an interface)?  Like this:
	e1 := e.(*TestGameEngine)

	if e0.id != e1.id {
		return false
	}
	if e0.name != e1.name {
		return false
	}
	if e0.actingPlayer != e1.actingPlayer {
		return false
	}
	if e0.actionCount != e1.actionCount {
		return false
	}
	if len(e0.players) != len(e1.players) {
		return false
	}
	for i, p0 := range e0.players {
		p1 := e1.players[i]
		if p0.Id != p1.Id {
			return false
		}
		if p0.Name != p1.Name {
			return false
		}
	}
	return true
}

func (g *TestGameEngine) Start(id string, name string, players []Player, seed int) GameState {
	_ = seed
	g.id = id
	g.name = name
	g.players = players

	return g.defaultGameState("Started game.", players[g.actingPlayer])
}

func (g *TestGameEngine) defaultGameState(text string, player Player) GameState {

	return GameState{
		Outcome:          Event{Text: text},
		ActingPlayer:     player,
		AvailableOptions: g.defaultOptions()}
}

func (g *TestGameEngine) defaultOptions() []Option {

	var options = make([]Option, 3)
	options[0] = Option{Abbr: "A", Text: "Act"}
	options[1] = Option{Abbr: "P", Text: "Pass"}
	options[2] = Option{Abbr: "Q", Text: "Quit"}

	return options
}

func (g *TestGameEngine) HandleAction(action Action) GameState {

	name := g.players[g.actingPlayer].Name
	if action.Abbr == "A" {
		g.actionCount += 1
		return GameState{
			Outcome:          Event{Text: fmt.Sprintf("%s acted.", name)},
			ActingPlayer:     g.players[g.actingPlayer],
			AvailableOptions: g.defaultOptions()}
	}
	if action.Abbr == "P" {
		g.actingPlayer += 1
		if g.actingPlayer >= len(g.players) {
			g.actingPlayer = 0
		}
		return GameState{
			Outcome:          Event{Text: fmt.Sprintf("%s passed.", name)},
			ActingPlayer:     g.players[g.actingPlayer],
			AvailableOptions: g.defaultOptions()}
	}
	if action.Abbr == "Q" {
		return GameState{
			Outcome: Event{Text: fmt.Sprintf("%s quit, game over.", name)}}
	}
	panic(fmt.Sprintf("Invalid abbr: %s", action.Abbr))
}

func (g *TestGameEngine) RefreshClient(_ string) string {
	return "RefreshClient."
}

func (g *TestGameEngine) Debug() string {
	return fmt.Sprintf("id=%s\nname=%s\nactionCount=%d\nactingPlayer=%d\nlen(g.players)=%d\n",
		g.id, g.name, g.actionCount, g.actingPlayer, len(g.players))
}
