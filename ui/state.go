// staacccccck
package ui

import (
	"github.com/76creates/stickers/flexbox"
	tea "github.com/charmbracelet/bubbletea"
)

type StateStack struct {
    states []AppState
}

// A view of the app
// not an interface unless we have view function?
// but don't use view function as we only call view
// on root model idk
type AppState interface {
    Update(msg tea.Msg) (AppState, tea.Cmd)
    RenderCells(*flexbox.FlexBox)
}

// Set state as the current app state
func (s *StateStack) Push(state AppState) {
    s.states = append(s.states, state)
}

// "Go back" - set the previous state of the app
// as the current
func (s *StateStack) Pop() AppState {
    if len(s.states) == 0 {
        return nil
    }

    top := s.states[len(s.states) - 1]
    s.states = s.states[:len(s.states) - 1]

    return top
}

// See the current state of the app
func (s *StateStack) Peek() AppState {
    if len(s.states) == 0 {
        return nil
    }

    return s.states[len(s.states) - 1]
}

func (s *StateStack) Length() int {
    return len(s.states)
}

