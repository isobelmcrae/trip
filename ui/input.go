package ui

import (
    "github.com/76creates/stickers/flexbox"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/isobelmcrae/trip/styles"
    "github.com/charmbracelet/log"
)

type inputState struct {
    root *RootModel
    input textinput.Model
    output *string
    
    // displayed to user
    prompt string
    placeholder string
}

func (s *inputState) Update(msg tea.Msg) (AppState, tea.Cmd){
    var cmd tea.Cmd
    s.input, cmd = s.input.Update(msg)

    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.Type == tea.KeyEnter {
            log.Debug("User input", "input", s.input.Value())
            s.root.States.Push(newSelectState(s.root, s.input.Value(), s.output))
            return s, cmd
        }
    }

    return s, cmd
}

// Update the main window and the sidebar's content
func (s *inputState) RenderCells(f *flexbox.FlexBox) {
    prompt := s.prompt
    sidebar := styles.WelcomeSidebarContent.Render(s.input.View())

    f.GetRow(0).GetCell(1).
        SetContent(styles.Prompt.Render(prompt) + "\n\n" + sidebar).
        SetStyle(styles.WelcomeSidebar)
}

// creates a new welcome state which can then
// be pushed onto states
func newInputState(root *RootModel, output *string, prompt string, placeholder string) AppState {
    ti := textinput.New()
    ti.Placeholder = placeholder
    ti.Focus()
    ti.Width = 30

    return &inputState{
        input: ti,
        root: root,

        output: output, 
        prompt: prompt,
        placeholder: placeholder,
    }
}


