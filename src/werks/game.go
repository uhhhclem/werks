package werks

import (
	"container/heap"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"uuid"
)

var LocosJsonPath = "../json/locos.json"

// Game represents a single game.
type Game struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Players      []*Player        `json:"players"`
	Locos        []*Loco          `json:"locos"`
	Turn         int              `json:"turn"`
	StartPlayer  int              `json:"startPlayer"`
	ActivePlayer int              `json:"currentPlayer"`
	Phase        Phase            `json:"phase"`
	Messages     Queue            `json:"-"`
	LocoMap      map[string]*Loco `json:"-"`
	TurnOrder	 	 PlayerQueue 			`json:"-"`
	PhaseOrder	 PlayerQueue 			`json:"-"`
}

// Player represents one of the players in the game.
type Player struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Money        int       `json:"money"`
	Factories    []Factory `json:"factories"`
	IsCurrent    bool      `json:"isCurrent"`
	TurnOrder    int       `json:"turnOrder"`
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
	Obsolete          bool   `json:"obsolete"`
}

// Die represents a slot where a die can be placed, and the die itself.
// If Render is false, no space will be rendered in the UI.  If Pips is
// 0, the die has no value.
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

var Phases = []string{
	"Locomotive Development",
	"Production Capacity",
	"Locomotive Production"}

// Actions represents the actions that are available to the
// current player.
type Actions struct {
	Phase string `json:"phase"`
	Actions []Action `json:"actions"`
}

// Action represents one action that is available to the active player.
type Action struct {
	Abbr string `json:"abbr"`
	Verb string `json:"verb"`
	Noun string `json:"noun"`
	Cost int `json:"cost"`
	Loco *Loco `json:"-"`
}

// GameState represents the current state of the game, including the
// currently available actions.
type GameState struct {
	Game *Game `json:"game"`
	Actions *Actions `json:"actions"`
}

func (g *Game) findAction(abbr string) *Action {
	actions := g.getActions()
	for _, a := range actions.Actions {
		if a.Abbr == abbr {
			return &a
		}
	}
	return nil
}

// performAction finds the Action for the given abbreviation and routes
// control to the handler appropriate for the phase.
func (g *Game) performAction(abbr string) {
	m := make(map[Phase]func(*Action))
	m[Development] = func(a *Action) { g.performDevelopmentAction(a)}
	m[Capacity] = func(a *Action) { g.performCapacityAction(a)}
	m[Production] = func(a *Action) { g.performProductionAction(a)}

	a := g.findAction(abbr)
	if a == nil {
		log.Printf("Couldn't find action for abbreviation %s", abbr)
	}

	m[g.Phase](a)
}

func (g *Game) performDevelopmentAction(a *Action) {
	// possible actions are:
	//   P = pass
	//   D = develop
	if strings.HasPrefix(a.Abbr, "D") {
		p := g.getCurrentPlayer()
		f := Factory {Key: a.Loco.Key, Capacity: 1}
		p.Factories = append(p.Factories, f)
		p.Money -= a.Loco.DevelopmentCost
	}
	p := g.getNextPlayer(false)
	if p == nil {
		g.Phase = Capacity
		g.setCurrentPlayer(g.getStartPlayer())
	} else {
		g.setCurrentPlayer(p)
	}
}

func (g *Game) performCapacityAction(a *Action) {
	log.Printf("Capacity action: %s", a.Abbr)
}

func (g *Game) performProductionAction(a *Action) {
	log.Printf("Production action: %s", a.Abbr)
}

// getActions returns the actions that are available to the current
// player right now.
func (g *Game) getActions() *Actions {

	actions := make([]Action, 0)

	if g.Phase == Development {
		for _, loco := range g.Locos {
			if g.isLocoAvailableForDevelopment(loco) {
				abbr := fmt.Sprintf("D:%s", loco.Key)
				actions = append(actions, Action{
					Abbr: abbr,
					Verb: "Develop",
					Noun: loco.Name,
					Cost: loco.DevelopmentCost,
					Loco: loco})
			}
		}
	}

	// I think you can always pass.
	actions = append(actions, Action{Abbr: "P", Verb: "Pass"})
	phase := Phases[g.Phase - 1]
	return &Actions{Phase: phase, Actions: actions}
}

// isLocoAvailableForDevelopment indicates if a given Loco is
// obsolete, has an Initial Orders die, or has at least one
// Existing Orders die.
func (g *Game) isLocoAvailableForDevelopment(loco *Loco) bool {
	if loco.Obsolete {
		return false
	}
	// if there's no initial order or existing order, loco can't
	// be developed.
	hasPips := loco.InitialOrders.Pips != 0
	for _, d := range loco.ExistingOrders {
		hasPips = hasPips || d.Pips != 0
	}
	if !hasPips {
		return false
	}
	// from here on, we can develop the loco unless the player can't.
	p := g.getCurrentPlayer()
	if loco.DevelopmentCost > p.Money {
		return false
	}
	for _, f := range p.Factories {
		if f.Key == loco.Key {
			return false
		}
	}
	return true
}

// makeNewGame creates a new game with the provided player names.
func makeNewGame(name string, playerNames []string) *Game {
	var g = new(Game)
	id, err := uuid.GenUUID()
	if err != nil {
		panic(err)
	}
	g.ID = id
	g.Name = name
	g.Messages.Capacity = 500
	g.Phase = Development
	Games[g.ID] = g
	g.addMessage(fmt.Sprintf("Created game %s...", g.Name))
	g.loadLocos()
	g.prepareLocos()
	g.initPlayers(playerNames)
	g.determineTurnOrder()

	return g
}


// getNextPlayer returns the next player.  If wrap is false, each player
// gets one turn, and this returns nil when all players have had a turn.
// If wrap is true, this will keep returning the next player in order
// indefinitely.
func (g *Game) getNextPlayer(wrap bool) *Player {
	if len(g.PhaseOrder) == 0 {
		if !wrap {
			return nil
		}
		g.PhaseOrder = make(PlayerQueue, len(g.Players))
		copy(g.PhaseOrder, g.TurnOrder)
	}
	pi := heap.Pop(&g.PhaseOrder).(*PlayerInfo)
	return pi.player
}

// determineTurnOrder initializes the TurnOrder priority
// queue, ordering players by money and previous turn
// order.  It also initializes the PhaseOrder queue.
func (g *Game) determineTurnOrder() {
	g.TurnOrder = make(PlayerQueue, 0, len(g.Players))
	for i, _ := range g.Players {
		pi := &PlayerInfo{player: g.Players[i], turnOrder: i}
		heap.Push(&g.TurnOrder, pi)
	}
	g.PhaseOrder = make(PlayerQueue, len(g.Players))
	copy(g.PhaseOrder, g.TurnOrder)

	// we have to copy the queue and pop everything out of
	// it to assign turn order to the players.
	var to = make(PlayerQueue, len(g.Players))
	copy(to, g.TurnOrder)

	for i, _ := range g.Players {
		pi := heap.Pop(&to).(*PlayerInfo)
		pi.player.TurnOrder = i
	}
}

// initPlayers sets up the Players array with initial values.
func (g *Game) initPlayers(names []string) {
	g.addMessage("Initializing players...")

	g.Players = make([]*Player, len(names))

	for i, name := range names {
		id, err := uuid.GenUUID()
		if err != nil {
			panic(err)
		}
		p := &Player{
			ID:           id,
			Name:         name,
			Factories:    make([]Factory, 1),
			Money:        12,
			ChatMessages: Queue{Capacity: 500},
			TurnOrder: 	  i}

		p.Factories[0] = Factory{Key: "p1", Capacity: 1}

		g.Players[i] = p
	}

	p := g.Players[0]
	p.IsCurrent = true
}

// getCurrentPlayer returns the current player.
func (g *Game) getCurrentPlayer() *Player {
	for _, p := range g.Players {
		if p.IsCurrent {
			return p
		}
	}
	panic("No current player!")
}

// getStartPlayer returns the current starting player.
func (g *Game) getStartPlayer() *Player {
	for _, p := range g.Players {
		if p.TurnOrder == 0 {
			return p
		}
	}
	panic("No start player!")
}

// setCurrentPlayer makes player p the active player.
func (g *Game) setCurrentPlayer(p *Player) {
	for i, p0 := range g.Players {
		if p0 == p {
			g.ActivePlayer = i
			p0.IsCurrent = true
		} else {
			p0.IsCurrent = false
		}
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
	g.LocoMap = make(map[string]*Loco)
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
		g.Locos[i].InitialOrders.Render = true
		g.Locos[i].ExistingOrders = make([]Die, 5)
		g.Locos[i].CustomerBase = make([]Die, 5)
		for j := 0; j < 5; j++ {
			g.Locos[i].ExistingOrders[j] = Die{Render: j < loco.MaxExistingOrders}
			g.Locos[i].CustomerBase[j] = Die{Render: j < loco.MaxCustomerBase}
		}
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

// getActionsJson marshals the available actions into a JSON byte slice.
func (g *Game) getActionsJson() []byte {
	actions := g.getActions()
	b, err := json.Marshal(actions)
	if err != nil {
		panic(err)
	}
	return b
}

// getGameJson marshals the Game into a JSON byte slice.
func (g *Game) getGameJson() []byte {
	b, err := json.Marshal(g)
	if err != nil {
		panic(err)
	}

	return b
}

func (g *Game) getGameStateJson() []byte {
	s := &GameState{Game: g, Actions: g.getActions()}
	b, err := json.Marshal(s)
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

var Games = make(map[string]*Game)
