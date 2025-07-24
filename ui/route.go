// FIX: assumes that there will always be a route available,
// crashes when there is no route
// TODO: display for when no routes are found
package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/isobelmcrae/trip/api"
	"github.com/isobelmcrae/trip/styles"
)

type legSelectionKeymap struct {
	PrevLeg key.Binding
	NextLeg key.Binding
}

// up to move up, down to move down
var legSelectionKeymapDefault = legSelectionKeymap{
	PrevLeg: key.NewBinding(key.WithKeys("pgup", "up", "j")),
	NextLeg: key.NewBinding(key.WithKeys("pgdown", "down", "k")),
}

type routeState struct {
	root           *RootModel
	Routes         []api.Journey
	paginator      paginator.Model
	legWidth       int
	loc            *time.Location

	legSelection int
}

// using only tfnsw for now
func (s *routeState) getRoutes() []api.Journey {
	// TODO: handle req which take a long time
	routes, err := s.root.Client.TripPlan(context.TODO(), s.root.OriginID, s.root.DestinationID)
	if err != nil {
		log.Debug("Error when fetching routes", "err", err)
	}

	if routes != nil {
		log.Debug("routes found", "count", len(routes))
	}

	return routes
}

func newRouteState(root *RootModel) AppState {
	width := root.flexBox.GetRow(0).GetCell(1).GetWidth() - 8
	location, _ := time.LoadLocation("Australia/Sydney")

	s := &routeState{
		root:     root,
		// TODO(iso): i am pretty sure this is the culprit here for non scaling/non
		legWidth: width,
		loc:      location,

		legSelection: 0,
	}

	originalRoutes := s.getRoutes()

	// extract styling to `styles/`
	s.paginator = paginator.New()
	s.paginator.Type = paginator.Dots
	s.paginator.PerPage = 1
	s.paginator.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).PaddingRight(1).Render("")
	s.paginator.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).PaddingRight(1).Render("")

	s.Routes = nil // append

	// format every route for pagination
	// ensure only routes after the current time are shown
	now := time.Now()
	for _, route := range originalRoutes {
		routeStartTime := route.Legs[0].Origin.DepartureTimeEstimated
		parsedTime, _ := time.Parse(time.RFC3339, routeStartTime)

		if parsedTime.After(now) {
			s.Routes = append(s.Routes, route)
		}
	}

	s.paginator.SetTotalPages(len(s.Routes))
	
	return s
}

// formats route-specific data for a given journey
func (s *routeState) displayRoute(r api.Journey) string {
	var doc strings.Builder

	origin := r.Legs[0].Origin
	destination := r.Legs[len(r.Legs)-1].Destination

	title := fmt.Sprintf("%s @%s\n\n%s @%s\n\n", origin.DisassembledName, formatTime(s.loc, origin.DepartureTimeEstimated), destination.DisassembledName, formatTime(s.loc, destination.ArrivalTimeEstimated))
	doc.WriteString(title)

	for idx, leg := range r.Legs {
		doc.WriteString(s.formatLeg(leg, idx))
	}

	return doc.String()
}

// format a time to sydney timezone
// TODO: proper error handling
func formatTime(loc *time.Location, rawTime string) string {
	if rawTime == "" {
		return "n/a"
	}

	parsed, _ := time.Parse(time.RFC3339, rawTime)
	formatted := parsed.In(loc).Format("3:04pm")
	return formatted
}

// formats the box for a given leg, takes in leg and its index into the Legs array
func (s *routeState) formatLeg(l api.Leg, idx int) string {
	var doc strings.Builder

	var transport string
	if l.Transportation.DisassembledName == "" {
		transport = "WALK"
	} else {
		transport = l.Transportation.DisassembledName
	}

	// colour highlight the line
	lineStr := styles.CreateLineHighlight(transport).Render(fmt.Sprintf("[%s]", transport))

	originStr := fmt.Sprintf("%s %s | %s", lineStr, l.Origin.DisassembledName, formatTime(s.loc, l.Origin.DepartureTimeEstimated))
	destStr := fmt.Sprintf("%s %s | %s", lineStr, l.Destination.DisassembledName, formatTime(s.loc, l.Destination.ArrivalTimeEstimated))

	duration := l.Duration / 60

	var showSelectedStr string
	isSelected := idx == s.legSelection
	if isSelected {
		showSelectedStr = " (focused)"
	}

	leg := fmt.Sprintf("%s\n\n> Travel for %dmin%s\n\n%s", originStr, duration, showSelectedStr, destStr)

	// bold iff. we have selected the leg
	doc.WriteString(styles.FormatRouteLeg(s.legWidth, transport, isSelected).Render(leg) + "\n")

	return doc.String()
}

func (s *routeState) RenderCells(f *flexbox.FlexBox) {
	arrowedPaginator := lipgloss.JoinHorizontal(lipgloss.Left, " ", s.paginator.View(), "")
	// styledRoute := styles.RouteTitle.Render(s.displayRoute(s.Routes[0]))
	styledPaginator := lipgloss.NewStyle().Width(s.legWidth).Align(lipgloss.Center).Render(arrowedPaginator)

	var b strings.Builder
	start, end := s.paginator.GetSliceBounds(len(s.Routes)) // returns (0, 1), (1, 2), (3, 4), etc
	for _, item := range s.Routes[start:end] {
		str := s.displayRoute(item)
		b.WriteString(str + "\n\n")
	}
	b.WriteString(styledPaginator)

	s.root.Sidebar.SetContent(b.String())

	// TODO render map here
	s.renderLeg(s.Routes[s.paginator.Page].Legs, s.legSelection)
}

func (s *routeState) Update(msg tea.Msg) (AppState, tea.Cmd) {
	var cmd tea.Cmd

	currentPage := s.paginator.Page
	s.paginator, cmd = s.paginator.Update(msg)
	if currentPage != s.paginator.Page {
		// reset legSelection
		s.legSelection = 0
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, legSelectionKeymapDefault.NextLeg):
			s.legSelection++
		case key.Matches(msg, legSelectionKeymapDefault.PrevLeg):
			s.legSelection--
		}
	}
	if s.legSelection >= len(s.Routes[s.paginator.Page].Legs) {
		s.legSelection = len(s.Routes[s.paginator.Page].Legs) - 1
	} else if s.legSelection < 0 {
		s.legSelection = 0
	}
	
	s.RenderCells(s.root.flexBox)

	return s, cmd
}
