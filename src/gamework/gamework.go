package gamework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Game is a single instance of a game.  The Id uniquely identifies the Game
// (typically it will be a uuid).  The Name is the game's human-readable name.
// State is the game's current state.  Actions contains a list of all of the
// Actions that have been taken in the game, in order.
type Game struct {
	Id      string
	Name    string
	Players []Player
	Seed    int
	Actions []Action
	State   GameState  `json:"-"`
	Engine  GameEngine `json:"-"`
}

// Player indicates a player in one specific Game.  The Id uniquely identifies
// the player within the game.  The Name (and/or the optional Detail, which by
// convention contains a JSON object) is used to present the player in the UI.
// The UserId uniquely identifies the user that owns the player.
//
// Note that Player does not contain a reference to the Game.  This is deliberate.
// A GameEngine may (and probably will, though it's not required to) maintain
// references to Players.  But it will never obtain a reference to the Game.
type Player struct {
	Id     string
	Name   string
	UserId string
	Detail *[]byte
}

// GameState indicates the game's current state.  Whenever the game changes state,
// the game will send the Outcome (an Event describing the transition to the new
// state, e.g. the result of the last action) to all players.  It will also send
// the AvailableOptions to the ActingPlayer.  If there are ever no AvailableOptions,
// the game is over.
//
// It's important to note that GameState is only a representation of the inputs
// that the game is currently waiting on.  The actual "state" of the game is
// internal to the GameEngine.  For example, the GameEngine of a chess game knows
// where all of the pieces are, but it only supports three or four GameStates:
// White's turn, Black's turn, and game over (or checkmate and stalemate).
type GameState struct {
	Outcome          Event
	ActingPlayer     Player
	AvailableOptions []Option
}

// Option is something that a Player can do right now.  The Abbr uniquely identifies
// the option; if the user chooses that option, the Abbr will be used to determine how
// to advance the game to the next state.  Text is a human-readable description of the
// option that is used when the game is running in console mode.  Detail is a payload
// of additional JSON-encoded data sent along with the Option to the client, with the
// expectation that the client will use that information in the presentation.
type Option struct {
	Abbr   string
	Text   string
	Detail *[]byte
}

// Action represents the action chosen by the user.  It will contain the Abbr of one
// of the chosen Options, and may additionally contain a payload of related JSON-encoded
// data in the Detail.
type Action struct {
	Abbr   string
	Detail *[]byte
}

// Event represents an occurrence in the game that should be presented to one ore
// more users.  Abbr identifies the type of Event, Text is a description of the Event
// (which more likely to be logged than simply presented to the user), and Detail is
// a payload of JSON-encoded related data.
type Event struct {
	Abbr   string
	Text   string
	Detail *[]byte
}

// GameEngine contains the actual logic of the game.
type GameEngine interface {

	// Start is used to initialize an instance of the engine for a new game.
	// It is expected (though not required) that the engine will maintain its
	// own copies of the game's ID, Name, and Players.  The seed is used to
	// initialize the random number generator.
	Start(id string, name string, players []Player, seed int) GameState

	// HandleAction is used to handle an action taken by the active player;
	// it returns the game's new state.
	HandleAction(action Action) GameState

	// RefreshClient is used to refresh the client with the game, typically
	// when the user hits F5 or reconnects to the server.  The string it returns
	// is a JSON payload that is sent to the client.
	RefreshClient(playerId string) string

	// Debug is another out-of-band interaction, used to dump debug information
	// to the console or browser.
	Debug() string

	// Equals is used primarily in testing:  it should return false if the two
	// engines aren't the same type, or if their internal states are unequal.
	Equals(e1 GameEngine) bool
}

// PlayToConsole allows testing and debugging of game engines without having to
// build a UI.
func PlayToConsole(g Game) {

	g.Actions = make([]Action, 0, 100)
	g.State = g.Engine.Start(g.Id, g.Name, g.Players, g.Seed)

	for {
		presentOptions(g.State)
		action := getAction(g.State)
		if action.Abbr == "" {
			s, err := WriteToString(g)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s\n", s)
			continue
		}
		g.State = performAction(&g, action)
		fmt.Printf("Outcome: %s\n", g.State.Outcome.Text)
		if len(g.State.AvailableOptions) == 0 {
			break
		}
	}
}

func performAction(g *Game, a Action) GameState {
	if g.Actions != nil {
		g.Actions = append(g.Actions, a)
	}
	return g.Engine.HandleAction(a)
}

func presentOptions(state GameState) {
	fmt.Printf("Acting player: %s\n\n", state.ActingPlayer.Name)
	for _, option := range state.AvailableOptions {
		fmt.Printf("%s: %s\n", option.Abbr, option.Text)
	}
	fmt.Printf("> ")
}

func getAction(state GameState) Action {
	var abbr string
	var action Action
	for {
		n, err := fmt.Scanln(&abbr)
		if err != nil {
			if n == 0 {
				break
			}
			panic(err)
		}
		if abbr == "" {
			break
		}
		for _, option := range state.AvailableOptions {
			if strings.ToUpper(abbr) == strings.ToUpper(option.Abbr) {
				action = Action{Abbr: option.Abbr}
				break
			}
		}
		if action.Abbr != "" {
			break
		}
		fmt.Printf("Invalid option: %s.\n", abbr)
		fmt.Printf("> ")
	}
	return action
}

// Replay takes a Game that has been deserialized (and has had its Engine
// assigned) and replays its stored actions to return the Engine to its
// current internal state.
func Replay(g Game) {
	g.State = g.Engine.Start(g.Id, g.Name, g.Players, g.Seed)
	for _, a := range g.Actions {
		g.State = g.Engine.HandleAction(a)
	}
}

// Serialize writes a JSON serialization of the game to the specified io.Writer.
func Serialize(g Game, w io.Writer) (n int, err error) {
	b, err := json.Marshal(g)
	if err != nil {
		return 0, err
	}
	return w.Write(b)
}

func Deserialize(r io.Reader, len int, g *Game) (err error) {
	b := make([]byte, len)
	_, err = r.Read(b)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, g)
}

func WriteToFile(g Game, filename string) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = Serialize(g, f)
	return err
}

func WriteToString(g Game) (s string, err error) {
	var b bytes.Buffer
	_, err = Serialize(g, &b)
	if err == nil {
		s = b.String()
	}
	return s, err
}

func ReadFromString(s string, g *Game) (err error) {
	b := bytes.NewBufferString(s)
	return Deserialize(b, b.Len(), g)
}
