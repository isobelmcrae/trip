package styles

import lg "github.com/charmbracelet/lipgloss"

// flexbox border styles
var (
    Border = lg.NewStyle().
        BorderStyle(lg.RoundedBorder())
)

// flexbox internal content styles
var (
    Prompt = lg.NewStyle().
        Bold(true).
        Align(lg.Center)
    WelcomeMain = lg.NewStyle().
        AlignHorizontal(lg.Center).
        AlignVertical(lg.Center).
        Align(lg.Center).
        Inherit(Border)
    WelcomeMainContent = lg.NewStyle().
        Align(lg.Center)
    WelcomeSidebar = lg.NewStyle().
        PaddingRight(2).
        PaddingLeft(2).
        PaddingTop(1).
        Inherit(Border)
    WelcomeSidebarContent = lg.NewStyle()
    LegBox = lg.NewStyle()
)
