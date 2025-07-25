package ui

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/isobelmcrae/trip/api"
	"github.com/isobelmcrae/trip/styles"
)

// routeState holds the state for the route view.
type routeState struct {
	root      *RootModel
	Routes    []api.Journey
	paginator paginator.Model
	viewport  viewport.Model
	legWidth  int
	loc       *time.Location
}

// getRoutes fetches trip plans from the API.
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

// newRouteState initializes the state for the route view.
func newRouteState(root *RootModel) AppState {
	location, _ := time.LoadLocation("Australia/Sydney")

	s := &routeState{
		root: root,
		loc:  location,
	}

	originalRoutes := s.getRoutes()

	// Filter routes to only include future journeys.
	now := time.Now()
	for _, route := range originalRoutes {
		if len(route.Legs) > 0 {
			routeStartTime := route.Legs[0].Origin.DepartureTimeEstimated
			parsedTime, err := time.Parse(time.RFC3339, routeStartTime)
			if err == nil && parsedTime.After(now) {
				s.Routes = append(s.Routes, route)
			}
		}
	}

	// measurements are relative to root's flexbox
	bigWidth := s.root.flexBox.GetWidth()
	width := int(math.Floor(float64(bigWidth)/10)*3) - 6
	height := s.root.flexBox.GetHeight() - 6

	// Update our leg width and the viewport's dimensions.
	s.legWidth = width - 2
	s.viewport.Width = width
	s.viewport.Height = height

	// Initialise the viewport.
	s.viewport = viewport.New(width, height)

	// Initialise the paginator.
	s.paginator = paginator.New()
	s.paginator.Type = paginator.Dots
	s.paginator.PerPage = 1
	s.paginator.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).PaddingRight(1).Render("⬤")
	s.paginator.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).PaddingRight(1).Render("⬤")
	s.paginator.SetTotalPages(len(s.Routes))

	// Set initial content, handling the no-routes case.
	if len(s.Routes) == 0 {
		s.viewport.SetContent(lipgloss.NewStyle().Width(s.legWidth).Align(lipgloss.Center).Render("No routes found."))
	} else {
		s.viewport.SetContent(s.displayRoute(s.Routes[0]))
	}

	return s
}

// displayRoute formats the details of a single journey into a string for the viewport.
func (s *routeState) displayRoute(r api.Journey) string {
	var doc strings.Builder

	if len(r.Legs) == 0 {
		return "This journey has no legs."
	}

	origin := r.Legs[0].Origin
	destination := r.Legs[len(r.Legs)-1].Destination

	title := fmt.Sprintf("%s @%s\n\n%s @%s\n\n", origin.DisassembledName, formatTime(s.loc, origin.DepartureTimeEstimated), destination.DisassembledName, formatTime(s.loc, destination.ArrivalTimeEstimated))
	doc.WriteString(title)

	for _, leg := range r.Legs {
		doc.WriteString(s.formatLeg(leg))
	}

	return doc.String()
}

// formatTime converts a time string to a readable format.
func formatTime(loc *time.Location, rawTime string) string {
	if rawTime == "" {
		return "n/a"
	}
	parsed, err := time.Parse(time.RFC3339, rawTime)
	if err != nil {
		return "invalid time"
	}
	return parsed.In(loc).Format("3:04pm")
}

// formatLeg formats the display for a single leg of a journey.
func (s *routeState) formatLeg(l api.Leg) string {
	var transport string
	if l.Transportation.DisassembledName == "" {
		transport = "WALK"
	} else {
		transport = l.Transportation.DisassembledName
	}

	lineStr := styles.CreateLineHighlight(transport).Render(fmt.Sprintf("[%s]", transport))
	originStr := fmt.Sprintf("%s %s | %s", lineStr, l.Origin.DisassembledName, formatTime(s.loc, l.Origin.DepartureTimeEstimated))
	destStr := fmt.Sprintf("%s %s | %s", lineStr, l.Destination.DisassembledName, formatTime(s.loc, l.Destination.ArrivalTimeEstimated))
	duration := l.Duration / 60
	leg := fmt.Sprintf("%s\n\n> Travel for %dmin\n\n%s", originStr, duration, destStr)

	// We pass false for isSelected because the viewport handles scrolling, not individual selection.
	return styles.FormatRouteLeg(s.legWidth, transport, false).Render(leg) + "\n"
}

// RenderCells renders the viewport and paginator into the flexbox layout.
func (s *routeState) RenderCells(f *flexbox.FlexBox) {
	var finalView string

	if len(s.Routes) > 0 {
		arrowedPaginator := lipgloss.JoinHorizontal(lipgloss.Left, "<", s.paginator.View(), ">")
		styledPaginator := lipgloss.NewStyle().Width(s.legWidth).Align(lipgloss.Center).Render(arrowedPaginator)
		finalView = lipgloss.JoinVertical(lipgloss.Left, s.viewport.View(), "\n", styledPaginator)
	} else {
		finalView = s.viewport.View()
	}

	s.root.Sidebar.SetContent(finalView)

	// TODO render map here
	if len(s.Routes) > 0 && s.paginator.Page < len(s.Routes) {
		// The second argument for renderLeg (selected leg) is 0 as we are not tracking it manually.
		s.renderLeg(s.Routes[s.paginator.Page].Legs, 0)
	}
}

// Update handles messages and updates the state.
func (s *routeState) Update(msg tea.Msg) (AppState, tea.Cmd) {
	var cmds []tea.Cmd
	var paginatorCmd, viewportCmd tea.Cmd

	pageBefore := s.paginator.Page

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// (TODO: move this and the init function's measurements code into a helper)
		// Get new dimensions
		bigWidth := s.root.flexBox.GetWidth()
		width := int(math.Floor(float64(bigWidth)/10)*3) - 6
		height := s.root.flexBox.GetHeight() - 6

		// Update dimensions
		s.legWidth = width - 2
		s.viewport.Width = width
		s.viewport.Height = height

		// Re-wrap the content in the viewport with the new width.
		if len(s.Routes) > 0 {
			s.viewport.SetContent(s.displayRoute(s.Routes[s.paginator.Page]))
		} else {
			s.viewport.SetContent(lipgloss.NewStyle().Width(s.legWidth).Align(lipgloss.Center).Render("No routes found."))
		}
		// We return nil here because we've handled this message and don't want to
		// pass it to the viewport's own update method, which would cause issues.
		return s, nil

	default:
		// For all other messages (like key presses), pass them to the children.
		if len(s.Routes) > 0 {
			s.paginator, paginatorCmd = s.paginator.Update(msg)
			s.viewport, viewportCmd = s.viewport.Update(msg)
			cmds = append(cmds, paginatorCmd, viewportCmd)
		}
	}

	// After any potential update, check if the page has changed.
	if len(s.Routes) > 0 && s.paginator.Page != pageBefore {
		s.viewport.SetContent(s.displayRoute(s.Routes[s.paginator.Page]))
		s.viewport.GotoTop()
	}

	s.RenderCells(s.root.flexBox)
	return s, tea.Batch(cmds...)
}
