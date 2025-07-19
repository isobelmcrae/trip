package styles

import lg "github.com/charmbracelet/lipgloss"

// flexbox border styles
var (
    InactiveStyle = lg.NewStyle().
        BorderStyle(lg.ThickBorder()).
        BorderForeground(InactiveColour)
    ActiveStyle = lg.NewStyle().
        BorderStyle(lg.ThickBorder()).
        BorderForeground(ActiveColour)
)

// flexbox internal content styles
var (
    WelcomeBox = lg.NewStyle().
        AlignHorizontal(lg.Center).
        AlignVertical(lg.Center)
    WelcomeBoxText = lg.NewStyle().
        Align(lg.Center)
    SidebarBox = lg.NewStyle().
        PaddingRight(2).
        PaddingLeft(2).
        PaddingTop(1)
    LegBox = lg.NewStyle() 
)
