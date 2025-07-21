package model

import (
	"database/sql"
	"log"
	"path"
	"runtime"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	trip "github.com/isobelmcrae/trip/api"

	_ "github.com/mattn/go-sqlite3"
)

// "states" - pages of the app
type AppState int 

const (
    StateTyping AppState = iota
    StateSelectingStop
    StateSelectingJourney
)

func forceRelativeToRoot(s string) string {
    _, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("unreachable")
	}
	return path.Join(path.Dir(filename), s)
}

var (
    // this is exceptionally shitty code, but it allows our tests to actually get the database from root
    DatabasePath = forceRelativeToRoot("../app.sqlite")
)

type Model struct {
    Db     *sql.DB // app.sqlite

    Client *trip.TripClient
    Width, Height int 

    TextInput textinput.Model
    StopList list.Model
    ActiveTabNum int
    
    Stops []trip.Stop
    Journeys []JourneyTab

    State AppState
    OriginID string
    DestinationID string 

    Keys keyMap
    Help help.Model
    Result string // maybe ditch
    Err error
}

func InitialModel() Model {
    ti := textinput.New()
    ti.Placeholder = "Enter origin stop..."
    ti.Focus()
    ti.CharLimit = 156
    ti.Width = 40

    sl := list.New([]list.Item{}, list.NewDefaultDelegate(), 20, 10)
    sl.Title = "Choose origin stop:"
    sl.SetShowStatusBar(false)

    db, err := sql.Open("sqlite3", DatabasePath)
    if err != nil {
        log.Fatal(err)
    }
    client := trip.NewClient(db)

    return Model{
        Client: client,
        TextInput: ti,
        StopList: sl,
        Help: help.New(),
        Keys: DefaultKeyMap(),
        State: StateTyping, // maybe create an title screen sorta thing
    }
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        textinput.Blink,
        tea.SetWindowTitle("trip planner :3"),
    )
}
