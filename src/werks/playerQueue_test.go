package werks

import (
	"container/heap"
	"testing"
)

func initPlayers() []Player {
	p := make([]Player, 4)
	p[0] = Player{Name: "Alpha", Money: 12}
	p[1] = Player{Name: "Beta", Money: 12}
	p[2] = Player{Name: "Gamma", Money: 12}
	p[3] = Player{Name: "Delta", Money: 12}
	return p
}

func TestPush(t *testing.T) {
	players := initPlayers()
	pq := make(PlayerQueue, 0, 4)
	for i, p := range players {
		heap.Push(&pq, &PlayerInfo{player: &p, turnOrder: i, index: i})
	}
	if pq.Len() != 4 {
		t.Errorf("pq.Len() was supposed to be 4.")
	}
}

func TestPop(t *testing.T) {
	players := initPlayers()
	players[0].Money = 30
	players[3].Money = 13

	pq := make(PlayerQueue, 0, 4)
	for i, _ := range players {
		pi := &PlayerInfo{player: &players[i], turnOrder: i}
		heap.Push(&pq, pi)
		t.Logf("Pushed %s (Money=%d, turnOrder=%d)\n",
			pi.player.Name, pi.player.Money, pi.turnOrder)
		t.Logf("Item %d of queue is %s", i, pq[i].player.Name)
	}
	for i := 0; i < 4; i++ {
		t.Logf("Item %d of queue is %s", i, pq[i].player.Name)
	}

	expectedOrder := []string {"Alpha", "Delta", "Beta", "Gamma"}
	for i, expected := range expectedOrder {
		pi := heap.Pop(&pq).(*PlayerInfo)
		t.Logf("Popped %s", pi.player.Name)
		if pi.player.Name != expected {
			t.Errorf("Expected item %d to be %s, but it's %s", i, expected, pi.player.Name)
		}
	}
}
