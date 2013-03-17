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
	"strings"
	"time"
	"user"
)

var rootPath string
var users *user.Users
var en = errors.New
var verboseLogs = false

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

type LoginResponse struct {
	Msg string `json:"msg"`
	Token string `json:"token"`
}

func invalidUserResponse(msg string) []byte {
	r := LoginResponse {Msg: msg}
	b, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return b
}

func validUserResponse(u *user.User) []byte {
	r := LoginResponse{Token: u.Token}
	b, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return b
}

// apiLoginHandler logs in a user.
func apiLoginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("u")
	password := r.FormValue("p")
	responseJson := invalidUserResponse("Invalid login.")
	if username != "" && password != "" {
		u, err := users.Login(username, password)
		if err == nil {
			responseJson = validUserResponse(u)
		}
	}
	w.Header().Add("content-type", "application/json")
	fmt.Fprintf(w, "%s", responseJson)
}

// apiRegisterHandler registers a user
func apiRegisterHandler (w http.ResponseWriter, r *http.Request) {
	var responseJson []byte
	var err error
	var u *user.User

	username := r.FormValue("u")
	password := r.FormValue("p")

	u, err = users.Register(username, password)
	if err != nil {
		responseJson = invalidUserResponse(fmt.Sprintf("%s", err))
	} else {
		err = users.SaveUsers()
		if err != nil {
			responseJson = invalidUserResponse("Registration failed.")
		} else {
			responseJson = validUserResponse(u)
		}
	}

	w.Header().Add("content-type", "application/json")
	fmt.Fprintf(w, "%s", responseJson)
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
		return;
	}

	if r.Method == "POST" {
		if err = r.ParseForm(); err != nil {
			serveError(w, err)
			return
		}

		abbr := r.FormValue("abbr")
		g.performAction(abbr)

		gameStateJson := g.getGameStateJson()
		w.Header().Add("content-type", "application/json")
		fmt.Fprintf(w, "%s", gameStateJson)
		return;
	}
}

// handleContentRequest handles most HTTP GET requests for static resources.
func handleContentRequest(w http.ResponseWriter, r *http.Request) {
	path := rootPath + r.URL.Path
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Couldn't open %s.", path)
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
	path = rootPath + "/html" + path
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Couldn't open %s.", path)
		http.NotFound(w, r)
		return
	}
	http.ServeContent(w, r, r.URL.Path, time.Unix(0, 0), file)
	file.Close()
}

func initApp() {
	rand.Seed(time.Now().UnixNano())
	users = user.Init("CaCl", "users.json")

	// load the users file, and create a default user if no users
	// file is found.
	var err error

	err = users.LoadUsers()
	if err != nil {
		_, err = users.Register("admin", "admin")
		if err != nil {
			panic(err)
		}
		err = users.SaveUsers()
		if err != nil {
			panic(err)
		}
	}
}

func Serve(path string) {
	rootPath = path

	initApp()

	// register handlers for the static URLs
	static_dirs := []string{"css", "html", "js", "lib", "views"}
	for _, path := range static_dirs {
		http.HandleFunc("/"+path+"/", handleContentRequest)
	}

	// register handlers for API calls
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/api/locos", apiLocosHandler)
	http.HandleFunc("/api/login", apiLoginHandler)
	http.HandleFunc("/api/register", apiRegisterHandler)
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
		if logUrl(r.URL.String()) {
			log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		}
		handler.ServeHTTP(w, r)
	})
}

func logUrl(url string) bool {
	if verboseLogs {
		return true
	}
	return !strings.HasPrefix(url, "/api/message") && !strings.HasPrefix(url, "/api/chat")
}
