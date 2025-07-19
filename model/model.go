package model

import (
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/help"
    "github.com/charmbracelet/bubbles/list"
    tea "github.com/charmbracelet/bubbletea"
    trip "github.com/isobelmcrae/trip/api"
)

// "states" - pages of the app
type AppState int 

const (
    StateTyping AppState = iota
    StateSelectingStop
    StateSelectingJourney
)

type Model struct {
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

    client := trip.NewClient("https://api.transport.nsw.gov.au/v1/tp")
    

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
