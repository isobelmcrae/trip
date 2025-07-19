package model

import "fmt"

func (m Model) View() string {
    switch m.State {
    case StateTyping:
        return fmt.Sprintf("Search stop:\n\n%s\n\n%s", m.TextInput.View(), m.Help.View(m.Keys))
    case StateSelectingStop:
        return fmt.Sprintf("%s\n\n", m.StopList.View())
    case StateSelectingJourney:
        return fmt.Sprintf("%s\n\n", showJourneys(m))
    default:
        return "Loading..."
    }
}
