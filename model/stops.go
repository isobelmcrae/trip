package model

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type stopItem struct {
    title string
    id    string
}

func (s stopItem) Title() string { 
    return s.title 
}
func (s stopItem) Description() string {
    return s.id
}
func (s stopItem) FilterValue() string {
    return s.title
}

// performs the search then formats for display in list
// changes state
func searchStops(m Model) (Model, tea.Cmd) {
    // returns ranked list of stops
    stops, err := m.Client.FindStop(m.TextInput.Value())
    if err != nil || len(stops) == 0 {
        m.Result = "No stops found" // FIX: handle properly
        return m, nil
    }

    if len(stops) == 1 && m.OriginID == "" {
        m.OriginID = stops[0].ID
        m.TextInput.SetValue("")
        m.TextInput.Placeholder = "Enter destination stop..."
        m.TextInput.Focus()
        m.State = StateTyping
        return m, nil
    }

    items := make([]list.Item, len(stops))
    for i, s := range stops {
        items[i] = stopItem{title: s.Name, id: s.ID}
    }
    
    if m.OriginID != "" { m.StopList.Title = "Select destination:" }
    m.StopList.SetItems(items)
    m.StopList.Select(0)
    m.State = StateSelectingStop

    return m, nil
}

func selectStop(m Model) (Model, tea.Cmd) {
    selected := m.StopList.SelectedItem().(stopItem)

    if m.OriginID == "" {
        m.OriginID = selected.id
        m.TextInput.SetValue("")
        m.TextInput.Placeholder = "Enter destination stop..."
        m.TextInput.Focus()
        m.State = StateTyping
    } else {
        m.DestinationID = selected.id
        return createJourneys(m)
    }
    return m, nil
}
