package ui

import (
    "database/sql"
    "github.com/76creates/stickers/flexbox"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/isobelmcrae/trip/styles"
    "github.com/isobelmcrae/trip/api"
    "github.com/isobelmcrae/trip/state"
    "github.com/charmbracelet/log"
    _ "github.com/mattn/go-sqlite3"
)

type RootModel struct {
    // each "state" modifies the content inside of
    // one of the flexbox cells
    flexBox *flexbox.FlexBox

    States StateStack

    client *api.TripClient

    OriginID string
    DestinationID string
}

func InitaliseRootModel() (m *RootModel){
    // create base flexbox cells
    m = &RootModel {
        flexBox: flexbox.New(0,0),
    }
    
    rows := []*flexbox.Row{
        m.flexBox.NewRow().AddCells(
            flexbox.NewCell(7, 1).SetStyle(styles.Border),
            flexbox.NewCell(3, 1).SetStyle(styles.Border),
        ),
    }

    m.flexBox.AddRows(rows)
    
    db, err := sql.Open("sqlite3", state.DatabasePath)
    if err != nil {
        log.Fatal(err)
    }
    // defer db.Close()
    m.client = api.NewClient(db)

    m.States.Push(newWelcomeState(m))

    return m
}

func (m *RootModel) Init() tea.Cmd {
    return tea.Batch(
        tea.SetWindowTitle("trip"),
    )
}

func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.flexBox.SetWidth(msg.Width)
        m.flexBox.SetHeight(msg.Height)
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyCtrlC:
            return m, tea.Quit
        }
    }

    current := m.States.Peek()
    if current == nil {
        return m, nil
    }
    
    oldStateSize := m.States.Size()
    updatedState, cmd := current.Update(msg)
    newStateSize := m.States.Size()

    if oldStateSize == newStateSize {
        m.States.states[len(m.States.states) - 1] = updatedState
    }

    m.View()

    return m, cmd
}

func (m *RootModel) View() string {
    state := m.States.Peek()
    if state != nil {
        state.RenderCells(m.flexBox)
    }
    return m.flexBox.Render()
}
