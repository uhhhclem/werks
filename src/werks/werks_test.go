package werks

import (
	"testing"
)

func TestPushMessage(t *testing.T) {
	var text string
	q := new(Queue)
	q.Capacity = 5

	// load up the queue to its capacity.
	tests := []string{"one", "two", "three", "four", "five"}
	for _, test := range tests {
		q.PushMessage(test)
	}

	text = q.Peek().Value.(*TextMessage).Text
	if text != "one" {
		t.Errorf("Tail should be one and is %s", text)
	}

	// pushing another message should throw the tail away
	q.PushMessage("six")
	text = q.Peek().Value.(*TextMessage).Text
	if text != "two" {
		t.Errorf("Tail should be two and is %s", text)
	}
}

func TestPopMessage(t *testing.T) {
	q := new(Queue)
	q.Capacity = 5
	tests := []string{"one", "two", "three", "four", "five"}
	for _, test := range tests {
		q.PushMessage(test)
	}
	for _, test := range tests {
		s := q.PopMessage()
		if s != test {
			t.Errorf("Got %s, expected %s.", s, test)
		}
	}
	if q.Count != 0 {
		t.Errorf("Queue should be empty, and isn't.")
	}
}

func newGame() *Game {
	name := "test"
	names := []string{"Abel", "Baker", "Charlie"}
	LocosJsonPath = "../../json/locos.json"
	g := makeNewGame(name, names)
	return g
}

func TestAddChatMessage(t *testing.T) {
	g := newGame()
	if g.Name != "test" {
		t.Errorf("Game has the wrong name.")
	}
	g.addChatMessage(g.Players[0], "test message")
	for _, p := range g.Players {
		t.Logf("Player %s", p.Name)
		c := g.getChatMessage(p)
		if c == nil {
			t.Errorf("Should have received at least one chat message.")
		}
		c = g.getChatMessage(p)
		if c != nil {
			t.Errorf("Should have received only one chat message.")
		}
	}
}

func TestIsLocoAvailableForDevelopment(t *testing.T) {
	g := newGame()
	// at start of the game, only a1 is available.
	for _, loco := range g.Locos {
		avail := g.isLocoAvailableForDevelopment(loco)
		if avail {
			if loco.Key != "a1" {
				t.Errorf("%s.InitialOrders.Render = %s", loco.Key, loco.InitialOrders.Render)
				t.Errorf("%s is available and shouldn't be", loco.Key)
			}
		} else {
			if loco.Key == "a1" {
				t.Errorf("%s isn't available and should be", loco.Key)
			}
		}
	}
}

func TestGetActions(t *testing.T) {
	g := newGame()
	if g.Phase != Development {
		t.Errorf("Game should start in the development phase.")
	}
	a := g.getActions()
	if a.Phase != "Locomotive Development" {
		t.Errorf("Game should start in the development phase.")
		return
	}
	actions := a.Actions
	expectedAbbrs := []string{"D:a1", "P"}
	if len(actions) != len(expectedAbbrs) {
		t.Errorf("There should be exactly %d actions at start.", len(expectedAbbrs))
		return
	}
	for i, expected := range expectedAbbrs {
		actual := actions[i].Abbr
		if expected != actual {
			t.Errorf("Expected %s, got %s", expected, actual)
		}
	}

}

func TestGetNextPlayer(t *testing.T) {
	var g *Game
	var p *Player
	var expectedNames []string

	t.Logf("wrap = true")
	g = newGame()
	expectedNames = []string{"Abel", "Baker", "Charlie", "Abel", "Baker", "Charlie"}
	for i, name := range expectedNames {
		p = g.getNextPlayer(true)
		if p.Name != name {
			t.Errorf("On iteration %d, expected %s and got %s", i, name, p.Name)
		}
	}

	t.Logf("wrap = false")
	g = newGame()
	expectedNames = []string{"Abel", "Baker", "Charlie"}
	for i, name := range expectedNames {
		p = g.getNextPlayer(false)
		if p.Name != name {
			t.Errorf("On iteration %d, expected %s and got %s", i, name, p.Name)
		}
	}
	p = g.getNextPlayer(false)
	if p != nil {
		t.Errorf("Expected to get nil, and got %s", p.Name)
	}

}
