package werks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"math/rand"
	"uuid"
)

var LocosJsonPath = "../json/locos.json"

// Game represents a single game.
type Game struct {
	ID           string    `json:"id"`
	Name string `json:"name"`
	Players      []*Player `json:"players"`
	Locos        []*Loco   `json:"locos"`
	Turn         int       `json:"turn"`
	StartPlayer  int       `json:"startPlayer"`
	ActivePlayer int       `json:"currentPlayer"`
	Phase        Phase     `json:"phase"`
	Messages     Queue     `json:"-"`
	LocoMap		   map[string] *Loco 	`json:"-"`
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
	Obsolete					bool	 `json:"obsolete"`
}

// Die represents a slot where a die can be placed, and the die itself.
// If Render is false, no space will be rendered in the UI.
type Die struct {
	Pips   int  `json:"pips"`
	Render bool `json:"render"`
}

// Phase identifies which phase is currently being executed.
type Phase int

const (
	Development Phase = iota + 1
	Capacity
	Production
)

// Action represents an action that is available to the active player.
type Action struct {
	Abbr	string	`json:"abbr"`
	Text 	string 	`json:"text"`
}

// getActions returns the actions that are available to the current
// user right now.
func (g *Game) getActions() []Action {
	result := make([]Action, 10)

	if g.Phase == Development {
		for _, loco := range g.Locos {
			if g.isLocoAvailableForDevelopment(loco) {
				abbr := fmt.Sprintf("D:%s", loco.Key)
				text := fmt.Sprintf("Develop %s for %d", loco.Name, loco.DevelopmentCost)
				result = append(result, Action{Abbr:abbr, Text:text})
			}
		}
	}

	// I think you can always pass.
	result = append(result, Action{Abbr:"P", Text:"Pass"})
	return result
}

// isLocoAvailableForDevelopment indicates if a given Loco is
// obsolete, has an Initial Orders die, or has at least one
// Existing Orders die.
func (g *Game) isLocoAvailableForDevelopment(loco *Loco) bool {
	if loco.Obsolete {
		return false
	}
	if loco.InitialOrders.Render {
		return true
	}
	for _, d := range loco.ExistingOrders {
		if d.Render {
			return true
		}
	}
	return false
}

// makeNewGame creates a new game with the provided player names.
func makeNewGame(name string, playerNames []string, testMode bool) *Game {
	var g = new(Game)
	id, err := uuid.GenUUID()
	if err != nil {
		panic(err)
	}
	g.ID = id
	g.Name = name
	g.Messages.Capacity = 500
	Games[g.ID] = g
	g.addMessage(fmt.Sprintf("Created game %s...", g.Name))
	g.loadLocos()
	g.prepareLocos()
	g.initPlayers(playerNames, testMode)
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
		id, err := uuid.GenUUID()
		if err != nil {
			panic(err)
		}
		g.Players[i] = &Player{
			ID:           id,
			Name:         name,
			Factories:    f,
			Money:        m,
			ChatMessages: Queue{Capacity: 500}}
	}

	p := g.Players[0]
	p.IsCurrent = true
	if testMode {
		text := fmt.Sprintf("Test message from %s", p.Name)
		g.addChatMessage(p, text)
	}
}

// rollDie rolls a Die and makes it visible.
func rollDie() Die {
	return Die{Pips: rand.Intn(6) + 1, Render: true}
}

// loadLocos loads (and unmarshals) Locos from their JSON representation.
func (g *Game) loadLocos() {
	path := LocosJsonPath
	result, err := ioutil.ReadFile(path)
	if err != nil {
		var pwd string
		pwd, err = os.Getwd()
		msg := fmt.Sprintf("reading file (pwd: %s, path: %s) failed", pwd, path)
		panic(msg)
	}

	var locos []Loco
	err = json.Unmarshal(result, &locos)
	g.Locos = make([]*Loco, len(locos))
	g.LocoMap = make(map[string] *Loco)
	for i, _ := range locos {
		g.Locos[i] = &locos[i]
		g.LocoMap[g.Locos[i].Key] = g.Locos[i]
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

// addMessage adds a new message to the game's queue.
func (g *Game) addMessage(text string) {
	g.Messages.PushMessage(text)
}

// addChatMessage adds a chat message to each player's queue.
func (g *Game) addChatMessage(player *Player, text string) {
	c := ChatMessage{Player: player, Text: text}
	for _, p := range g.Players {
		p.ChatMessages.Push(c)
	}
}

// getMessage gets the next message from the game's queue, or the empty string.
func (g *Game) getMessage() string {
	if g.Messages.Count == 0 {
		return ""
	}
	return g.Messages.PopMessage()
}

// getChatMessage gets the next chat message from the player's queue, or nil.
func (g *Game) getChatMessage(player *Player) *ChatMessage {
	e := player.ChatMessages.Pop()
	if e == nil {
		return nil
	}
	c := e.Value.(ChatMessage)
	return &c
}

// pushChatMessage pushes a message from a player into all players' queues.
func (g *Game) pushChatMessage(player *Player, text string) {
		m := ChatMessage{Player: player, Text: text}
		for _, p := range g.Players {
			p.ChatMessages.Push(m)
		}
}

// getPlayer returns the player with the specified ID
func (g *Game) getPlayer(id string) (*Player, error) {
	for _, p := range g.Players {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, errors.New(
		fmt.Sprintf("Unknown player ID: %s", id))
}

var Phases = []string{
	"Locomotive Development",
	"Production Capacity",
	"Locomotive Production"}

var Games = make(map[string]*Game)
