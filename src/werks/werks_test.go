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

func TestInitPlayers(t *testing.T) {
	name := "test"
	names := []string{"Abel", "Baker", "Charlie"}
	LocosJsonPath = "../../json/locos.json"
	g := makeNewGame(name, names, true)
	if g.Name != name {
		t.Errorf("Game has the wrong name.")
	}
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
