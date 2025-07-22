package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/isobelmcrae/trip/ui"
	"github.com/joho/godotenv"
	"os" 
)

func main() {
    // log to file
    f, err := os.OpenFile("trip.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
	panic(err)
    }
    log.SetOutput(f)
    log.SetLevel(log.DebugLevel)
    log.SetReportCaller(true)

    errenv := godotenv.Load()
    if errenv != nil {
        log.Error("Error loading .env")
    }

    m := ui.InitaliseRootModel()
    p := tea.NewProgram(m, tea.WithAltScreen())

    if _, err := p.Run(); err != nil {
        log.Fatal("TUI error:", err)
    }
}
