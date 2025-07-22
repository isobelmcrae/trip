package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/isobelmcrae/trip/ui"
	"github.com/joho/godotenv"
)

func main() {
    log.SetLevel(log.DebugLevel)
    err := godotenv.Load()
    if err != nil {
        log.Error("Error loading .env")
    }

    m := ui.InitaliseRootModel()
    p := tea.NewProgram(m, tea.WithAltScreen())

    if _, err := p.Run(); err != nil {
        log.Fatal("TUI error:", err)
    }
}
