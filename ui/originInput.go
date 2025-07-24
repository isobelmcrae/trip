package ui

import (
    "github.com/76creates/stickers/flexbox"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/isobelmcrae/trip/styles"
    "github.com/charmbracelet/log"
)

type originInputState struct {
    root *RootModel
    input textinput.Model
    output *string
}

func (s *originInputState) Update(msg tea.Msg) (AppState, tea.Cmd){
    var cmd tea.Cmd
    s.input, cmd = s.input.Update(msg)

    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.Type == tea.KeyEnter {
            log.Debug("User origin input", "input", s.input.Value())
            s.root.States.Push(newOriginSelectState(s.root, s.input.Value()))
            return s, cmd
        }
    }

    return s, cmd
}

// Update the main window and the sidebar's content
func (s *originInputState) RenderCells(f *flexbox.FlexBox) {
    sidebar := styles.WelcomeSidebarContent.Render(s.input.View())

    f.GetRow(0).GetCell(1).
        SetContent(styles.Prompt.Render("Where are you?") + "\n\n" + sidebar).
        SetStyle(styles.WelcomeSidebar)
}

// creates a new welcome state which can then
// be pushed onto states
func newOriginInputState(root *RootModel) AppState {
    ti := textinput.New()
    ti.Placeholder = "Enter origin stop..."
    ti.Focus()
    ti.Width = 30

    return &originInputState{
        input: ti,
        root: root,
    }
}



