package model

import (
    "fmt"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    trip "github.com/isobelmcrae/trip/api"

    "strings"
    lg "github.com/charmbracelet/lipgloss"
)

// https://github.com/charmbracelet/bubbletea/tree/main/examples/tabs
type JourneyTab struct {
    Title string
    Arrival time.Time
    Departure time.Time
    Duration int
    Details trip.Journey
}

// styling for tab menu
var (
    inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
    activeTabBorder = tabBorderWithBottom("┘", " ", "└")
    docStyle = lg.NewStyle().Padding(1, 2, 1, 2)
    highlightColor = lg.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
    inactiveTabStyle = lg.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1).Foreground(lg.Color("#a3a3a3"))
    activeTabStyle = inactiveTabStyle.Border(activeTabBorder, true).Bold(true).Foreground(lg.Color("#ffffff"))
    windowStyle = lg.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Border(lg.NormalBorder()).UnsetBorderTop()
)

func createJourneys(m Model) (Model, tea.Cmd) {
    sydneyLocation, _ := time.LoadLocation("Australia/Sydney")
    journeys, err := m.Client.TripPlan(m.OriginID, m.DestinationID)
    if err != nil || len(journeys) == 0 {
        // need to display that no routes were found
        return m, nil
    }

    var tabs []JourneyTab
    
    // create the tab menu + title for each
    for _, j := range journeys {
        var tab = JourneyTab{}
        tab.Details = j

        var duration int
        for _, leg := range j.Legs {
            duration += leg.Duration
        }

        tab.Duration = duration

        startRaw := j.Legs[0].Origin.DepartureTimeEstimated
        endRaw := j.Legs[len(j.Legs)-1].Destination.ArrivalTimeEstimated

        start, _ := time.Parse(time.RFC3339, startRaw)
        depStr := start.In(sydneyLocation).Format("15:04")
        end, _ := time.Parse(time.RFC3339, endRaw)
        arrStr := end.In(sydneyLocation).Format("15:04")

        tab.Departure = start
        tab.Arrival = end

        tab.Title = fmt.Sprintf("%s -> %s", depStr, arrStr)

        tabs = append(tabs, tab)
    }
    
    m.Journeys = tabs
    m.State = StateSelectingJourney
    m.ActiveTabNum = 0

    return m, nil
}

func showJourneys(m Model) string {
    if len(m.Journeys) == 0 {
        return "No journeys"
    }

    var doc strings.Builder
    var renderedTabs []string
    
    // format the tabs (activetab vs inactive)
    for i, j := range m.Journeys {
        style := inactiveTabStyle
        isFirst := i == 0
        isLast := i == len(m.Journeys) - 1
        isActive := i == m.ActiveTabNum

        if isActive {
            style = activeTabStyle
        }

        border, _, _, _, _ := style.GetBorder()
        if isFirst && isActive {
            border.BottomLeft = "│"
        } else if isFirst {
            border.BottomLeft = "├"
        }

        if isLast && isActive {
            border.BottomRight = "│"
        } else if isLast {
            border.BottomRight = "┤"
        }
        style = style.Border(border)

        renderedTabs = append(renderedTabs, style.Render(j.Title))
    }

    activeTab := m.Journeys[m.ActiveTabNum]
    activeTabContent := formatJourney(activeTab.Details)

    tabs := lg.JoinHorizontal(lg.Top, renderedTabs...)
    content := windowStyle.Width(lg.Width(tabs) - windowStyle.GetHorizontalFrameSize()).Render(
        activeTabContent,
    )

    doc.WriteString(tabs + "\n")
    doc.WriteString(content)

    return docStyle.Render(doc.String())
}

func formatJourney (journey trip.Journey) string {
    var content strings.Builder

    // title
    title := fmt.Sprintf("%s -> %s", journey.Legs[0].Origin.DisassembledName, journey.Legs[len(journey.Legs) - 1].Destination.DisassembledName)

    content.WriteString(title + "\n")

    // display each leg:
    // origin, destination, transportation, duration
    // start and end time
    content.WriteString("Legs:\n")
    for i, leg := range journey.Legs {
        leg := fmt.Sprintf("\nLeg %d: %s (time here) -> %s (time here)\n%s | %d for %dmin", i + 1, leg.Origin.DisassembledName, leg.Destination.DisassembledName, leg.Transportation.Number, leg.Distance, leg.Duration / 60)
        content.WriteString(leg + "\n")
    }

    return content.String()
}

// create border
func tabBorderWithBottom(left, middle, right string) lg.Border {
    border := lg.RoundedBorder()
    border.BottomLeft = left
    border.Bottom = middle
    border.BottomRight = right

    return border
}

