package gamework

import (
	"strings"
	"testing"
)

// makeAction finds the option matching abbr, and creates an Action
// out of it.
func makeAction(abbr string, options []Option) Action {
	for _, o := range options {
		if strings.ToUpper(o.Abbr) == strings.ToUpper(abbr) {
			return Action{Abbr: o.Abbr}
		}
	}
	return Action{}
}

// playGame pretends to play our game for a few turns.
func playGame(g *Game) {

	var a Action

	for i := 0; i < 3; i++ {
		a = makeAction("P", g.State.AvailableOptions)
		g.State = performAction(g, a)
	}
	a = makeAction("A", g.State.AvailableOptions)
	g.State = performAction(g, a)

	a = makeAction("P", g.State.AvailableOptions)
	g.State = performAction(g, a)

}

func TestSerializationAndDeserialization(t *testing.T) {
	// First, let's set up a game, assign it an engine, and play
	// a couple of turns.

	g0 := InitTestGameWithTestEngine()
	e0 := g0.Engine.(*TestGameEngine)
	g0.Actions = make([]Action, 0, 100)
	g0.State = g0.Engine.Start(g0.Id, g0.Name, g0.Players, 0)

	playGame(&g0)

	// Now serialize the game to s0.
	s0, err := WriteToString(g0)
	if err != nil {
		t.Errorf("%s", err)
		return
	}

	// Create a new game g1 and deserialize it from s0.  Note that
	// g1 doesn't have an engine yet.
	var g1 Game
	var s1 string
	err = ReadFromString(s0, &g1)
	if err != nil {
		t.Errorf("%s", err)
	}

	// If we serialize g1 to s1, s1 and s0 should be equal.
	s1, err = WriteToString(g1)
	if s0 != s1 {
		t.Errorf("Serialization failed.\n")
		t.Errorf("s0 = %s", s0)
		t.Errorf("s1 = %s", s1)
	}

	// Create an engine for the game, and replay the stored Actions.
	e1 := new(TestGameEngine)
	g1.Engine = e1
	Replay(g1)

	// The two engines should have the same internal state.
	if !e0.Equals(e1) {
		t.Errorf("Game engines aren't equal.\n")
		t.Errorf("\ng0.Engine.Debug() = \n%s", e0.Debug())
		t.Errorf("\ng1.Engine.Debug() = \n%s", e1.Debug())
	}
}
