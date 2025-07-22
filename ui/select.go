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
    output *string
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
            // create and push next state IF there are stops
            selectedItem := s.selectionList.SelectedItem()
            if selectedItem != nil {
                selectedID := selectedItem.(stopItem).id
                log.Debug("stop selected", "id", selectedID)
                *s.output = selectedID

                // push next thingy to stack w/ the other param
                // NOTE: this is not adaptable to other use cases beyond
                // selecting origin/destination - idk how to adapt
                if s.root.DestinationID == "" {
                    s.root.States.Push(newInputState(s.root, &s.root.DestinationID, "Where are you going?", "Enter destination stop..."))
                } else {
                    s.root.States.Push(newRouteState(s.root))
                }
            }
            return s, cmd
        }
    
    }

    return s, cmd
}

// updates sidebar flexbox to display the selection list
func (s *selectState) RenderCells(f *flexbox.FlexBox) {
    prompt := "Select stop:\n"

    // TODO: better way to store these values?
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
func newSelectState(root *RootModel, input string, output *string) AppState {
    sl := list.New([]list.Item{}, list.NewDefaultDelegate(), 20, 10)
    sl.SetShowTitle(false)
    sl.SetShowHelp(false)
    sl.SetShowStatusBar(false)
    
    m := &selectState{
        selectionList: sl,
        root: root,
        input: input,
        output: output,
    }

    selectStop(m)

    return m
}

// gets the stops which match the input string, formats them into a selection list
func selectStop(m *selectState) {
    stops := m.root.Client.FindStop(m.input)
    if len(stops) == 0 {
        log.Debug("No stops found")
    }

    listItems := make([]list.Item, len(stops))
    for i, stop := range stops {
        listItems[i] = stopItem{ title: stop.Name, id: stop.ID }
    }

    m.selectionList.SetItems(listItems)
    m.selectionList.Select(0)
}
