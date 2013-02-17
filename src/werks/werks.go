package werks

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var en = errors.New

// getGameFromRequest finds the game whose ID is in the URL's query string or form.
func getGameFromRequest(r *http.Request) (*Game, error) {
	id := r.FormValue("g")
	g, ok := Games[id]
	if !ok {
		return nil, errors.New("Unknown game ID: " + id)
	}
	return g, nil
}

// getGameAndPlayerFromRequest finds the game and player from the request
func getGameAndPlayerFromRequest(r *http.Request) (*Game, *Player, error) {
	var g *Game
	var p *Player
	var ok bool
	var err error

	g_id := r.FormValue("g")
	p_id := r.FormValue("p")
	g, ok = Games[g_id]
	if !ok {
		return nil, nil, errors.New("Unrecognized game ID.")
	}
	p, err = g.getPlayer(p_id)
	if err != nil {
		return nil, nil, err
	}
	return g, p, nil
}

// apiGameHandler returns the requested game.
func apiGameHandler(w http.ResponseWriter, r *http.Request) {
	g, err := getGameFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	gameJson := g.getGameJson()
	w.Header().Add("content-type", "application/json")
	fmt.Fprintf(w, "%s", gameJson)
}

// apiLocosHandler returns the Locos in the requested game.
func apiLocosHandler(w http.ResponseWriter, r *http.Request) {
	g, err := getGameFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	locoJson := g.getLocoJson()
	w.Header().Add("content-type", "application/json")
	fmt.Fprintf(w, "%s", locoJson)
}

// apiPlayersHandler returns the Players in the requested game.
func apiPlayersHandler(w http.ResponseWriter, r *http.Request) {
	g, err := getGameFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	playersJson := g.getPlayersJson()
	w.Header().Add("content-type", "application/json")
	fmt.Fprintf(w, "%s", playersJson)
}

func apiMessageHandler(w http.ResponseWriter, r *http.Request) {
	g, err := getGameFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	messageJson := g.getMessageJson()
	if messageJson == nil {
		return
	}
	w.Header().Add("content-type", "application/json")
	fmt.Fprintf(w, "%s", messageJson)
}

type ChatMessageJson struct {
	Who  string `json:"who"`
	Text string `json:"text"`
}

func apiChatHandler(w http.ResponseWriter, r *http.Request) {
	g, p, err := getGameAndPlayerFromRequest(r)
	if err != nil {
		serveError(w, err)
		return
	}
	if r.Method == "GET" {
		e := p.ChatMessages.Pop()
		if e == nil {
			return
		}
		m := e.Value.(ChatMessage)
		c := ChatMessageJson{Who: m.Player.Name, Text: m.Text}
		b, err := json.Marshal(c)
		if err != nil {
			panic(err)
		}
		w.Header().Add("content-type", "application/json")
		fmt.Fprintf(w, "%s", b)
	}
	if r.Method == "POST" {
		text := r.FormValue("text")
		g.pushChatMessage(p, text)
	}
}

// serveError writes an error message.
func serveError(w http.ResponseWriter, e error) {
	http.Error(w, e.Error(), http.StatusInternalServerError)
}

// apiNewGameHandler creates a new game.
func apiNewGameHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	if r.Method != "POST" {
		return
	}
	if err = r.ParseForm(); err != nil {
		serveError(w, err)
		return
	}

	var playerCount int
	if playerCount, err = strconv.Atoi(r.FormValue("playerCount")); err != nil {
		serveError(w, err)
	}

	name := r.FormValue("name")
	playerNames := make([]string, playerCount)
	for i := 0; i < playerCount; i++ {
		key := fmt.Sprintf("player%d", i)
		playerNames[i] = r.FormValue(key)
	}

	g := makeNewGame(name, playerNames)
	gameJson := g.getGameJson()
	w.Header().Add("content-type", "application/json")
	fmt.Fprintf(w, "%s", gameJson)
}

func apiActionHandler(w http.ResponseWriter, r *http.Request) {
	var g *Game
	var p *Player
	var err error

	g, p, err = getGameAndPlayerFromRequest(r)
	if err != nil {
		serveError(w, err)
	}
	if r.Method == "GET" {
		if !p.IsCurrent {
			return
		}
		actionsJson := g.getActionsJson()
		w.Header().Add("content-type", "application/json")
		fmt.Fprintf(w, "%s", actionsJson)
	}

	if r.Method != "POST" {
		return
	}
	if err = r.ParseForm(); err != nil {
		serveError(w, err)
		return
	}

	abbr := r.FormValue("abbr")

	fmt.Printf("%s", abbr)
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

func initApp() {
	rand.Seed(time.Now().UnixNano())
}

func Serve() {
	initApp()

	// register handlers for the static URLs
	static_dirs := []string{"css", "html", "js", "lib", "views"}
	for _, path := range static_dirs {
		http.HandleFunc("/"+path+"/", handleContentRequest)
	}

	// register handlers for API calls
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/api/locos", apiLocosHandler)
	http.HandleFunc("/api/players", apiPlayersHandler)
	http.HandleFunc("/api/game", apiGameHandler)
	http.HandleFunc("/api/newGame", apiNewGameHandler)
	http.HandleFunc("/api/message", apiMessageHandler)
	http.HandleFunc("/api/chat", apiChatHandler)
	http.HandleFunc("/api/action", apiActionHandler)

	// start serving
	http.ListenAndServe(":8080", Log(http.DefaultServeMux))
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
