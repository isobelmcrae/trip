package ui

import (
	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
        "github.com/charmbracelet/log"
	// "github.com/charmbracelet/lipgloss"
	"github.com/isobelmcrae/trip/styles"
)


type selectState struct {
    root *RootModel
    selectionList list.Model
    input string
}

type stopItem struct {
    title string
    id string
}

func (sI stopItem) Title() string {
    return sI.title
}

func (sI stopItem) Description() string {
    return sI.id
}

func (sI stopItem) FilterValue() string {
    return sI.title
}

func (s *selectState) Update(msg tea.Msg) (AppState, tea.Cmd){
    var cmd tea.Cmd

    s.selectionList, cmd = s.selectionList.Update(msg)

    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.Type == tea.KeyEnter {
            // create and push next state
            // selected := s.selectionList.SelectedItem().(stopItem).id
            return s, tea.Quit
        }
    }

    return s, cmd
}

func (s *selectState) RenderCells(f *flexbox.FlexBox) {
    prompt := "Select a starting stop:\n"
    sidebarHeight := f.GetRow(0).GetCell(1).GetHeight()
    sidebarWidth := f.GetRow(0).GetCell(1).GetWidth()

    s.selectionList.SetSize(sidebarWidth - 7, sidebarHeight - 10)

    sidebar := styles.WelcomeSidebarContent.Render(styles.Prompt.Render(prompt) + s.selectionList.View())

    f.GetRow(0).GetCell(1).
        SetContent(sidebar).
        SetStyle(styles.WelcomeSidebar)
}

// creates a new selection state which can then
// be pushed onto states
func newSelectState(root *RootModel, input string) AppState {
    sl := list.New([]list.Item{}, list.NewDefaultDelegate(), 20, 10)
    sl.SetShowTitle(false)
    sl.SetShowHelp(false)
    sl.SetShowStatusBar(false)
    
    m := &selectState{
        selectionList: sl,
        root: root,
        input: input,
    }

    selectStop(m)

    return m
}

// gets the stops which match the input string, formats them into a selection list
func selectStop(m *selectState) {
    stops := m.root.client.FindStop(m.input)
    if len(stops) == 0 {
        log.Debug("No stops found.")
    }

    listItems := make([]list.Item, len(stops))
    for i, stop := range stops {
        listItems[i] = stopItem{ title: stop.Name, id: stop.ID }
    }

    m.selectionList.SetItems(listItems)
    m.selectionList.Select(0)
}
