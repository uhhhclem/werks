package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
  "net/http"
  "os"
	"strconv"
  "time"
)

var en = errors.New

// Game represents a single game.
type Game struct {
	ID string `json:"id"`
	Players []*Player `json:"players"`
	Locos []*Loco `json:"locos"`
}

// Player represents one of the players in the game.
type Player struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Money int `json:"money"`
	Factories []Factory `json:"factories"`
	IsCurrent bool `json:"isCurrent"`
}

// Factory represents a factory owned by a player.
type Factory struct {
	Key string `json:"key"`
	Capacity int `json:"capacity"`
	UnitsSold int `json:"unitsSold"`
}

// Loco represents a locomotive type.  Its attributes are a superset
// of the board state and the information on the card.
type Loco struct {
	Kind string `json:"kind"`
	Generation int `json:"generation"`
	Name string `json:"name"`
	Years string `json:"years"`
	DevelopmentCost int `json:"developmentCost"`
	ProductionCost int `json:"productionCost"`
	Income int `json:"income"`
	MaxExistingOrders int `json:"maxExistingOrders"`
	MaxCustomerBase int `json:"maxCustomerBase"`
	Count int `json:"count"`
	Key string `json:"key"`
	UpgradeTo string `json:"upgradeTo"`
	UpgradeCost int `json:"upgradeCost"`
	ExistingOrders []Die `json:"existingOrders"`
	InitialOrders Die `json:"initialOrders"`
	CustomerBase []Die `json:"customerBase"`
}

// Die represents a slot where a die can be placed, and the die itself.
// If Render is false, no space will be rendered in the UI.
type Die struct {
	Pips int `json:"pips"`
	Render bool `json:"render"`
}

var Games = make(map[string] *Game)
var LastGameID int

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
		for j:=0; j < 5; j++ {
			g.Locos[i].ExistingOrders[j] = Die {Render: j < loco.MaxExistingOrders}
			g.Locos[i].CustomerBase[j] = Die {Render: j < loco.MaxCustomerBase }
		}
		g.Locos[i].InitialOrders = Die {Render: true}
		nextLoco := g.findUpgrade(loco)
		if nextLoco != nil {
			key = fmt.Sprintf("%s%d", prefixes[loco.Kind], loco.Generation + 1)
			g.Locos[i].UpgradeTo = key
			g.Locos[i].UpgradeCost = nextLoco.ProductionCost - loco.ProductionCost
		}
	}
	g.Locos[0].ExistingOrders[0] = rollDie();
	g.Locos[0].ExistingOrders[1] = rollDie();
	g.Locos[0].ExistingOrders[2] = rollDie();

	g.Locos[1].InitialOrders = rollDie();
}

// findUpgrade finds the Loco (if any) to upgrade oldLoco to.
func (g *Game) findUpgrade(oldLoco *Loco) *Loco {
	for _, newLoco := range g.Locos {
		if oldLoco.Kind == newLoco.Kind && oldLoco.Generation + 1 == newLoco.Generation {
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

func (g *Game) getPlayersJson() []byte {
	b, err := json.Marshal(g.Players)
	if err != nil {
		panic(err)
	}

	return b
}

func rollDie() Die {
  return Die {Pips: rand.Intn(6) + 1, Render: true}
}

// getGameFromURL finds the game whose ID is in the URL's query string.
func getGameFromURL(r *http.Request) (*Game, error) {
	id := r.FormValue("g")
	g, ok := Games[id]
	if !ok {
		return nil, errors.New("Unknown game ID: " + id)
	}
	return g, nil
}

func apiGameHandler(w http.ResponseWriter, r *http.Request) {
	g , err := getGameFromURL(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	gameJson := g.getGameJson()
	w.Header().Add("content-type", "application/json")
	fmt.Fprintf(w, "%s", gameJson)
}

func apiLocosHandler(w http.ResponseWriter, r *http.Request) {
	g , err := getGameFromURL(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	locoJson := g.getLocoJson()
	w.Header().Add("content-type", "application/json")
	fmt.Fprintf(w, "%s", locoJson)
}

func apiPlayersHandler(w http.ResponseWriter, r *http.Request) {
	g , err := getGameFromURL(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	playersJson := g.getPlayersJson()
	w.Header().Add("content-type", "application/json")
	fmt.Fprintf(w, "%s", playersJson)
}

func serveError(w http.ResponseWriter, e error) {
	http.Error(w, e.Error(), http.StatusInternalServerError)
}

func apiNewGameHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	if r.Method != "POST" {
		return;
	}
	if err = r.ParseForm(); err != nil {
		serveError(w, err)
		return
	}

	var playerCount int
	if playerCount, err = strconv.Atoi(r.FormValue("playerCount")); err != nil {
		serveError(w, err)
	}

	names := make([]string, playerCount)
	for i := 0; i < playerCount; i++ {
		key := fmt.Sprintf("player%d", i)
		names[i] = r.FormValue(key)
	}

	g := makeNewGame(names)
	gameJson := g.getGameJson()
	w.Header().Add("content-type", "application/json")
	fmt.Fprintf(w, "%s", gameJson)
}

func apiActionEndPhaseHandler(w http.ResponseWriter, r *http.Request) {

}

// handleContentRequest handles most HTTP GET requests for static resources.
func handleContentRequest(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("../" + r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeContent(w, r, r.URL.Path, time.Unix(0, 0), file)
	file.Close()
}

// rootHandler handles requests from the root URL
func rootHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}
	file, err := os.Open("../html" + path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeContent(w, r, r.URL.Path, time.Unix(0, 0), file)
	file.Close()
}

// initPlayers sets up the Players array with initial values.
func (g *Game) initPlayers(names []string, testMode bool) {
	var f []Factory
	var m int

	g.Players = make([]*Player, len(names))

	if testMode {
		f = makeTestFactories()
		m = 30
	} else {
		f = makeStandardFactories()
		m = 12
	}

	for i, name := range names {
		id := fmt.Sprintf("%d", i)
		g.Players[i] = &Player {ID: id, Name: name, Factories: f, Money: m}
	}
	g.Players[0].IsCurrent = true
}

// makeTestFactories returns a test slice of Factories.
func makeTestFactories() []Factory {
	f := make([]Factory, 3)
	f[0] = Factory {Key: "p1", Capacity:1 }
	f[1] = Factory {Key: "a1", Capacity:1 }
	f[2] = Factory {Key: "p2", Capacity:0 }
	return f
}

// makeStandardFactories returns the slice of Factories set up per the rules.
func makeStandardFactories() []Factory {
	f := make([]Factory, 1)
	f[0] = Factory {Key: "p1", Capacity:1 }
	return f
}

func initApp() {
		rand.Seed(time.Now().UnixNano())
}

// makeNewGame creates a new game, currently with the default player
// names.
func makeNewGame(names []string) *Game {
	LastGameID += 1
	id := fmt.Sprintf("%d", LastGameID)
	var g = new(Game)
	g.ID = id
	Games[g.ID] = g
	g.loadLocos()
	g.prepareLocos()
	g.initPlayers(names, true)
	return g
}

func main() {
	initApp()

	// register handlers for the static URLs
	static_dirs := []string { "css", "html", "js", "lib", "views"}
	for _, path := range static_dirs {
		http.HandleFunc("/" + path + "/", handleContentRequest)
	}

	// register handlers for API calls
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/api/locos", apiLocosHandler)
	http.HandleFunc("/api/players", apiPlayersHandler)
	http.HandleFunc("/api/game", apiGameHandler)
	http.HandleFunc("/api/newGame", apiNewGameHandler)
	http.HandleFunc("/api/action/endPhase", apiActionEndPhaseHandler);

	// start serving
	http.ListenAndServe(":8080", nil)
}
