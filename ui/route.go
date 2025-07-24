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
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/isobelmcrae/trip/api"
	"github.com/isobelmcrae/trip/styles"
	"github.com/mitchellh/go-wordwrap"
)

type routeState struct {
	root           *RootModel
	Routes         []api.Journey
	RenderedRoutes []string
	paginator      paginator.Model
	legWidth       int
	loc            *time.Location
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
		legWidth: width,
		loc:      location,
	}

	s.Routes = s.getRoutes()

	// extract styling to `styles/`
	s.paginator = paginator.New()
	s.paginator.Type = paginator.Dots
	s.paginator.PerPage = 1
	s.paginator.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).PaddingRight(1).Render("")
	s.paginator.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).PaddingRight(1).Render("")

	// format every route for pagination
	// ensure only routes after the current time are shown
	var formattedRoutes []string
	now := time.Now()
	for _, route := range s.Routes {
		routeStartTime := route.Legs[0].Origin.DepartureTimeEstimated
		log.Debug("start time:", "time", routeStartTime)
		parsedTime, _ := time.Parse(time.RFC3339, routeStartTime)
		log.Debug("time parsed:", "parsed", parsedTime)

		if parsedTime.After(now) {
			formatted := s.displayRoute(route)
			formattedRoutes = append(formattedRoutes, formatted)
		}
	}

	s.RenderedRoutes = formattedRoutes
	s.paginator.SetTotalPages(len(s.RenderedRoutes))
	
	return s
}

// formats route-specific data for a given journey
func (s *routeState) displayRoute(r api.Journey) string {
	var doc strings.Builder

	origin := r.Legs[0].Origin
	destination := r.Legs[len(r.Legs)-1].Destination

	title := fmt.Sprintf("%s @%s\n\n%s @%s\n\n", origin.DisassembledName, formatTime(s.loc, origin.DepartureTimeEstimated), destination.DisassembledName, formatTime(s.loc, destination.ArrivalTimeEstimated))
	doc.WriteString(title)

	for _, leg := range r.Legs {
		doc.WriteString(s.formatLeg(leg))
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

// formats the box for a given leg
func (s *routeState) formatLeg(l api.Leg) string {
	var doc strings.Builder

	var transport string
	if l.Transportation.DisassembledName == "" {
		transport = "WALK"
	} else {
		transport = l.Transportation.DisassembledName
	}

	// colour highlight the line
	lineStr := styles.CreateLineHighlight(transport).Render(fmt.Sprintf("[%s]", transport))

	originStr := fmt.Sprintf("%s %s\n", lineStr, l.Origin.DisassembledName)
	destStr := fmt.Sprintf("%s %s", lineStr, l.Destination.DisassembledName)

	// wrap to fit the flexbox
	wrappedOStr := wordwrap.WrapString(originStr, uint(s.legWidth))
	wrappedDStr := wordwrap.WrapString(destStr, uint(s.legWidth))

	leg := fmt.Sprintf("%s\n%s", wrappedOStr, wrappedDStr)

	doc.WriteString(styles.FormatRouteLeg(s.legWidth, transport).Render(leg) + "\n")

	return doc.String()
}

func (s *routeState) RenderCells(f *flexbox.FlexBox) {
	arrowedPaginator := lipgloss.JoinHorizontal(lipgloss.Left, " ", s.paginator.View(), "")
	// styledRoute := styles.RouteTitle.Render(s.displayRoute(s.Routes[0]))
	styledPaginator := lipgloss.NewStyle().Width(s.legWidth).Align(lipgloss.Center).Render(arrowedPaginator)

	var b strings.Builder
	start, end := s.paginator.GetSliceBounds(len(s.RenderedRoutes))
	for _, item := range s.RenderedRoutes[start:end] {
		b.WriteString(item + "\n\n")
	}
	b.WriteString(styledPaginator)

	s.root.Sidebar.SetContent(b.String())

	// TODO render map here
	s.renderLeg(s.Routes[0].Legs, 0)
}

func (s *routeState) Update(msg tea.Msg) (AppState, tea.Cmd) {
	var cmd tea.Cmd

	s.paginator, cmd = s.paginator.Update(msg)

	return s, cmd
}
