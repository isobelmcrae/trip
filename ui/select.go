package ui

import (
	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	// "github.com/charmbracelet/lipgloss"
	"github.com/isobelmcrae/trip/styles"
)


type selectState struct {
    root *RootModel
    selectionList list.Model
}

func (s *selectState) Update(msg tea.Msg) (AppState, tea.Cmd){
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.Type == tea.KeyEnter {
            // create and push next state
            return s, tea.Quit
        }
    }

    return s, cmd
}

func (s *selectState) RenderCells(f *flexbox.FlexBox) {
    sidebar := styles.WelcomeSidebarContent.Render("this is going to be a selection list")
    main := styles.WelcomeMainContent.Render(welcome)

    f.GetRow(0).GetCell(1).
        SetContent(sidebar).
        SetStyle(styles.WelcomeSidebar)
    f.GetRow(0).GetCell(0).SetContent(main).
        SetStyle(styles.WelcomeMain)
}

// creates a new selection state which can then
// be pushed onto states
func newSelectState(root *RootModel) AppState {
    sl := list.New([]list.Item{}, list.NewDefaultDelegate(), 20, 10)
    sl.Title = "Choose origin stop:"
    
    return &selectState{
        selectionList: sl,
        root: root,
    }
}
