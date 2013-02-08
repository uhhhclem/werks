package werks

import (
	"errors"
	"fmt"
	"math/rand"
  "net/http"
  "os"
	"strconv"
  "time"
)

var en = errors.New

// getGameFromURL finds the game whose ID is in the URL's query string.
func getGameFromURL(r *http.Request) (*Game, error) {
	id := r.FormValue("g")
	g, ok := Games[id]
	if !ok {
		return nil, errors.New("Unknown game ID: " + id)
	}
	return g, nil
}

// apiGameHandler returns the requested game.
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

// apiLocosHandler returns the Locos in the requested game.
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

// apiPlayersHandler returns the Players in the requested game.
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

func apiMessageHandler(w http.ResponseWriter, r *http.Request) {
	g , err := getGameFromURL(r)
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

// serveError writes an error message.
func serveError(w http.ResponseWriter, e error) {
	http.Error(w, e.Error(), http.StatusInternalServerError)
}

// apiNewGameHandler creates a new game.
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

/*
The possible player actions (and the phase they can occur in):

Phase  Action
-----  ------
  1    Develop locomotive %key.
  1    Pass.
  2    Buy %amt capacity on factory %key.
  2    Upgrade %amt capacity from factory %key.
  2    Pass.
  3    Sell from factory %key.
  3    Pass.
*/
func apiPlayerActionHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	if r.Method != "POST" {
		return;
	}
	if err = r.ParseForm(); err != nil {
		serveError(w, err)
		return
	}

	action := r.FormValue("action")
	key := r.FormValue("key")
	amt := r.FormValue("amt")

	fmt.Printf("%s %s %s", action, key, amt)
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


func Serve() {
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
	http.HandleFunc("/api/message", apiMessageHandler)

	// start serving
	http.ListenAndServe(":8080", nil)
}
