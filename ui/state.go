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
// Prevents user from showing a state before the welcome
func (s *StateStack) Pop() AppState {
    if len(s.states) == 1 {
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

// see how many states currently are on the stack
func (s *StateStack) Size() int {
    return len(s.states)
}

