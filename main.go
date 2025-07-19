package main

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/joho/godotenv"
    "github.com/isobelmcrae/trip/model"
    "github.com/charmbracelet/log"
)

func main() {
    log.SetLevel(log.DebugLevel)
    err := godotenv.Load()
    if err != nil {
        log.Error("Error loading .env")
    }

    m := model.InitialModel()
    p := tea.NewProgram(m, tea.WithAltScreen())

    if _, err := p.Run(); err != nil {
        log.Fatal("TUI error:", err)
    }
}
