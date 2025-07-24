package ui

import (
	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
        "github.com/charmbracelet/log"
	// "github.com/charmbracelet/lipgloss"
	"github.com/isobelmcrae/trip/styles"
)


type destSelectState struct {
    root *RootModel
    selectionList list.Model
    listSize int
    input string
}

type destStopItem struct {
    title string
    id string
}

func (sI destStopItem) Title() string {
    return sI.title
}

func (sI destStopItem) Description() string {
    return sI.id
}

func (sI destStopItem) FilterValue() string {
    return sI.title
}

func (s *destSelectState) Update(msg tea.Msg) (AppState, tea.Cmd){
    var cmd tea.Cmd

    s.selectionList, cmd = s.selectionList.Update(msg)

    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.Type == tea.KeyEnter {
            // create and push next state IF there are stops
            if s.listSize == 0 {
                s.root.States.Pop()
                return s, cmd
            }

            selectedItem := s.selectionList.SelectedItem()

            selectedID := selectedItem.(destStopItem).id
            log.Debug("stop selected", "id", selectedID)
            s.root.DestinationID = selectedID

            s.root.States.Push(newRouteState(s.root))

            return s, cmd
        }
    
    }

    return s, cmd
}

// updates sidebar flexbox to display the selection list
func (s *destSelectState) RenderCells(f *flexbox.FlexBox) {
    prompt := "Select stop:\n"

    // TODO: better way to store these values?
    sidebarHeight := s.root.Sidebar.GetHeight()
    sidebarWidth := s.root.Sidebar.GetWidth()

    s.selectionList.SetSize(sidebarWidth - 7, sidebarHeight - 10)

    sidebar := styles.WelcomeSidebarContent.Render(styles.Prompt.Render(prompt) + s.selectionList.View())

    f.GetRow(0).GetCell(1).
        SetContent(sidebar).
        SetStyle(styles.WelcomeSidebar)
}

// creates a new selection state which can then
// be pushed onto states
func newDestSelectState(root *RootModel, input string) AppState {
    sl := list.New([]list.Item{}, list.NewDefaultDelegate(), 20, 10)
    sl.SetShowTitle(false)
    sl.SetShowHelp(false)
    sl.SetShowStatusBar(false)
    
    m := &destSelectState{
        selectionList: sl,
        root: root,
        input: input,
    }

    destSelectStop(m)

    return m
}

// gets the stops which match the input string, formats them into a selection list
func destSelectStop(m *destSelectState) {
    stops := m.root.Client.FindStop(m.input)

    m.listSize = len(stops)
    if m.listSize == 0 {
        log.Debug("No stops found")
    }

    listItems := make([]list.Item, len(stops))
    for i, stop := range stops {
        listItems[i] = destStopItem{ title: stop.Name, id: stop.ID }
    }

    m.selectionList.SetItems(listItems)
    m.selectionList.Select(0)
}
