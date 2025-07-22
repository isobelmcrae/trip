package ui

import (
    "github.com/76creates/stickers/flexbox"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/isobelmcrae/trip/styles"
)

const welcome = "trip v0.0.0\n\nsydney public transport for your terminal\n\nhjkl/arrow keys to move\nenter to select"

type welcomeState struct {
    root *RootModel
    input textinput.Model
}

func (s *welcomeState) Update(msg tea.Msg) (AppState, tea.Cmd){
    var cmd tea.Cmd
    s.input, cmd = s.input.Update(msg)

    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.Type == tea.KeyEnter {
            s.root.States.Push(newSelectState(s.root, s.input.Value()))
            return s, cmd
        }
    }

    return s, cmd
}

// Update the main window and the sidebar's content
func (s *welcomeState) RenderCells(f *flexbox.FlexBox) {
    prompt := "Where are you?"
    sidebar := styles.WelcomeSidebarContent.Render(s.input.View())
    main := styles.WelcomeMainContent.Render(welcome)

    f.GetRow(0).GetCell(1).
        SetContent(styles.Prompt.Render(prompt) + "\n\n" + sidebar).
        SetStyle(styles.WelcomeSidebar)
    f.GetRow(0).GetCell(0).SetContent(main).
        SetStyle(styles.WelcomeMain)
}

// creates a new welcome state which can then
// be pushed onto states
func newWelcomeState(root *RootModel) AppState {
    ti := textinput.New()
    ti.Placeholder = "Enter origin stop..."
    ti.Focus()
    ti.Width = 30

    return &welcomeState{
        input: ti,
        root: root,
    }
}


