package ui

import (
    "github.com/76creates/stickers/flexbox"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/isobelmcrae/trip/styles"
    "github.com/charmbracelet/log"
)

type destInputState struct {
    root *RootModel
    input textinput.Model
}

func (s *destInputState) Update(msg tea.Msg) (AppState, tea.Cmd){
    var cmd tea.Cmd
    s.input, cmd = s.input.Update(msg)

    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.Type == tea.KeyEnter {
            log.Debug("User input", "input", s.input.Value())
            s.root.States.Push(newDestSelectState(s.root, s.input.Value()))
            return s, cmd
        }
    }

    return s, cmd
}

// Update the main window and the sidebar's content
func (s *destInputState) RenderCells(f *flexbox.FlexBox) {
    sidebar := styles.WelcomeSidebarContent.Render(s.input.View())

    f.GetRow(0).GetCell(1).
        SetContent(styles.Prompt.Render("Where are you going?") + "\n\n" + sidebar).
        SetStyle(styles.WelcomeSidebar)
}

// creates a new welcome state which can then
// be pushed onto states
func newDestInputState(root *RootModel) AppState {
    ti := textinput.New()
    ti.Placeholder = "Enter destination stop..."
    ti.Focus()
    ti.Width = 30

    return &destInputState{
        input: ti,
        root: root,
    }
}



