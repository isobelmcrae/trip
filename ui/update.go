package model

import (
    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    
    // handle key input
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // custom keys for text input
        if m.TextInput.Focused() {
            m.TextInput, cmd = m.TextInput.Update(msg)
            if msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC {
                return m, tea.Quit
            }
        }

        switch {
        case key.Matches(msg, m.Keys.Quit):
            return m, tea.Quit
        case key.Matches(msg, m.Keys.Help):
            m.Help.ShowAll = !m.Help.ShowAll
        }

    case tea.WindowSizeMsg:
        m.Width = msg.Width
        m.Height = msg.Height
        m.StopList.SetSize(m.Width, m.Height - 7)
        m.Help.Width = msg.Width 
    }
    
    // handle state
    m, cmd = handleState(m, cmd, msg)
    return m, cmd
}

func handleState(m Model, cmd tea.Cmd, msg tea.Msg) (Model, tea.Cmd) {
    switch m.State {
    case StateTyping:
        if msg, ok := msg.(tea.KeyMsg); ok && key.Matches(msg, m.Keys.Select) {
            return searchStops(m)
        }
        return m, cmd
    
    case StateSelectingStop:
        if msg, ok := msg.(tea.KeyMsg); ok && key.Matches(msg, m.Keys.Select) {
            return selectStop(m)
        }
        m.StopList, cmd = m.StopList.Update(msg)
        return m, cmd
    
    case StateSelectingJourney:
        if msg, ok := msg.(tea.KeyMsg); ok {
            switch {
            case key.Matches(msg, m.Keys.Left):
                m.ActiveTabNum = max(0, m.ActiveTabNum - 1)
            case key.Matches(msg, m.Keys.Right):
                m.ActiveTabNum = min(len(m.Journeys) - 1, m.ActiveTabNum + 1)
            }
        }
        return m, cmd
    }

    return m, nil
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
