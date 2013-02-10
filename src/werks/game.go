package werks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
)

// Game represents a single game.
type Game struct {
	ID           string    `json:"id"`
	Players      []*Player `json:"players"`
	Locos        []*Loco   `json:"locos"`
	Turn         int       `json:"turn"`
	StartPlayer  int       `json:"startPlayer"`
	ActivePlayer int       `json:"currentPlayer"`
	Phase        int       `json:"phase"`
	Messages     Queue     `json:"-"`
}

// Player represents one of the players in the game.
type Player struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Money        int       `json:"money"`
	Factories    []Factory `json:"factories"`
	IsCurrent    bool      `json:"isCurrent"`
	ChatMessages Queue     `json:"-"`
}

// Factory represents a factory owned by a player.
type Factory struct {
	Key       string `json:"key"`
	Capacity  int    `json:"capacity"`
	UnitsSold int    `json:"unitsSold"`
}

// Loco represents a locomotive type.  Its attributes are a superset
// of the board state and the information on the card.
type Loco struct {
	Kind              string `json:"kind"`
	Generation        int    `json:"generation"`
	Name              string `json:"name"`
	Years             string `json:"years"`
	DevelopmentCost   int    `json:"developmentCost"`
	ProductionCost    int    `json:"productionCost"`
	Income            int    `json:"income"`
	MaxExistingOrders int    `json:"maxExistingOrders"`
	MaxCustomerBase   int    `json:"maxCustomerBase"`
	Count             int    `json:"count"`
	Key               string `json:"key"`
	UpgradeTo         string `json:"upgradeTo"`
	UpgradeCost       int    `json:"upgradeCost"`
	ExistingOrders    []Die  `json:"existingOrders"`
	InitialOrders     Die    `json:"initialOrders"`
	CustomerBase      []Die  `json:"customerBase"`
}

// Die represents a slot where a die can be placed, and the die itself.
// If Render is false, no space will be rendered in the UI.
type Die struct {
	Pips   int  `json:"pips"`
	Render bool `json:"render"`
}

// makeNewGame creates a new game with the provided player names.
func makeNewGame(names []string) *Game {
	LastGameID += 1
	id := fmt.Sprintf("%d", LastGameID)
	var g = new(Game)
	g.ID = id
	g.Messages.Capacity = 500
	Games[g.ID] = g
	g.addMessage(fmt.Sprintf("Created game %s...", g.ID))
	g.loadLocos()
	g.prepareLocos()
	g.initPlayers(names, true)
	return g
}

// initPlayers sets up the Players array with initial values.
func (g *Game) initPlayers(names []string, testMode bool) {
	g.addMessage("Initializing players...")
	var f []Factory
	var m int

	g.Players = make([]*Player, len(names))

	if testMode {
		g.addMessage("Running in test mode...")
		f = makeTestFactories()
		m = 30
	} else {
		f = makeStandardFactories()
		m = 12
	}

	for i, name := range names {
		id := fmt.Sprintf("%d", i)
		g.Players[i] = &Player{
			ID:        id,
			Name:      name,
			Factories: f,
			Money:     m}
	}
	g.Players[0].IsCurrent = true
}

// rollDie rolls a Die and makes it visible.
func rollDie() Die {
	return Die{Pips: rand.Intn(6) + 1, Render: true}
}

// loadLocos loads (and unmarshals) Locos from their JSON representation.
func (g *Game) loadLocos() {
	path := "../json/locos.json"
	result, err := ioutil.ReadFile(path)
	if err != nil {
		panic(path + " unreadable.")
	}

	var locos []Loco
	err = json.Unmarshal(result, &locos)
	g.Locos = make([]*Loco, len(locos))
	for i, _ := range locos {
		g.Locos[i] = &locos[i]
	}
	if err != nil {
		panic(err)
	}
}

// prepareLocos assigns some values that are implicit in the data but it's
// useful to precompute:  key, upgradeTo, upgradeCost.
func (g *Game) prepareLocos() {
	g.addMessage("Preparing locomotives...")
	prefixes := make(map[string]string, 5)
	prefixes["passenger"] = "p"
	prefixes["fast"] = "a"
	prefixes["freight"] = "g"
	prefixes["special"] = "s"

	for i, loco := range g.Locos {
		key := fmt.Sprintf("%s%d", prefixes[loco.Kind], loco.Generation)
		g.Locos[i].Key = key
		g.Locos[i].ExistingOrders = make([]Die, 5)
		g.Locos[i].CustomerBase = make([]Die, 5)
		for j := 0; j < 5; j++ {
			g.Locos[i].ExistingOrders[j] = Die{Render: j < loco.MaxExistingOrders}
			g.Locos[i].CustomerBase[j] = Die{Render: j < loco.MaxCustomerBase}
		}
		g.Locos[i].InitialOrders = Die{Render: true}
		nextLoco := g.findUpgrade(loco)
		if nextLoco != nil {
			key = fmt.Sprintf("%s%d", prefixes[loco.Kind], loco.Generation+1)
			g.Locos[i].UpgradeTo = key
			g.Locos[i].UpgradeCost = nextLoco.ProductionCost - loco.ProductionCost
		}
	}
	g.Locos[0].ExistingOrders[0] = rollDie()
	g.Locos[0].ExistingOrders[1] = rollDie()
	g.Locos[0].ExistingOrders[2] = rollDie()

	g.Locos[1].InitialOrders = rollDie()
}

// findUpgrade finds the Loco (if any) to upgrade oldLoco to.
func (g *Game) findUpgrade(oldLoco *Loco) *Loco {
	for _, newLoco := range g.Locos {
		if oldLoco.Kind == newLoco.Kind && oldLoco.Generation+1 == newLoco.Generation {
			return newLoco
		}
	}
	return nil
}

// getGameJson marshals the Game into a JSON byte slice.
func (g *Game) getGameJson() []byte {
	b, err := json.Marshal(g)
	if err != nil {
		panic(err)
	}

	return b
}

// getLocoJson marshals the Locos into a JSON byte slice.
func (g *Game) getLocoJson() []byte {
	b, err := json.Marshal(g.Locos)
	if err != nil {
		panic(err)
	}

	return b
}

// getPlayersJson marshals the Players into a JSON byte slice.
func (g *Game) getPlayersJson() []byte {
	b, err := json.Marshal(g.Players)
	if err != nil {
		panic(err)
	}

	return b
}

// getPlayersJson marshals the next Message into a JSON byte slice
func (g *Game) getMessageJson() []byte {
	e := g.Messages.Pop()
	if e == nil {
		return nil
	}
	m := e.Value.(*TextMessage)
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	return b
}

// addMessage adds a new message to the game's queue
func (g *Game) addMessage(text string) {
	g.Messages.PushMessage(text)
}

// getMessage gets the next message from the game's queue, or the empty string.
func (g *Game) getMessage() string {
	if g.Messages.Count == 0 {
		return ""
	}
	return g.Messages.PopMessage()
}

var Phases = []string{
	"Locomotive Development",
	"Production Capacity",
	"Locomotive Production"}

var Games = make(map[string]*Game)

var LastGameID int
