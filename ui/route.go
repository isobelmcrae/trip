package ui

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/viewport"
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
	PrevLeg: key.NewBinding(key.WithKeys("up", "j")),
	NextLeg: key.NewBinding(key.WithKeys("down", "k")),
}

// routeState holds the state for the route view.
type routeState struct {
	root         *RootModel
	Routes       []api.Journey
	paginator    paginator.Model
	viewport     viewport.Model
	legWidth     int
	loc          *time.Location
	legSelection int
	legOffsets   []int // vertical positions of each leg
	legHeights   []int // actual heights of each leg
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
		root:         root,
		loc:          location,
		legSelection: 0,
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
		s.setViewportContent(0)
	}

	return s
}

// setViewportContent sets the viewport content and calculates leg offsets
func (s *routeState) setViewportContent(routeIndex int) {
	if routeIndex >= len(s.Routes) {
		return
	}

	content, offsets, heights := s.displayRouteWithOffsetsAndHeights(s.Routes[routeIndex])
	s.legOffsets = offsets
	s.legHeights = heights
	s.viewport.SetContent(content)
}

// scrollToSelectedLeg scrolls the viewport to keep the selected leg visible
func (s *routeState) scrollToSelectedLeg() {
	if s.legSelection >= len(s.legOffsets) || len(s.legOffsets) == 0 {
		return
	}

	selectedLegOffset := s.legOffsets[s.legSelection]
	selectedLegHeight := s.legHeights[s.legSelection]
	viewportTop := s.viewport.YOffset
	viewportBottom := viewportTop + s.viewport.Height

	// Special case: if first leg is selected, scroll to top to show header
	if s.legSelection == 0 {
		s.viewport.GotoTop()
		return
	}

	// If selected leg is above the viewport, scroll up to show it
	if selectedLegOffset < viewportTop {
		s.viewport.SetYOffset(selectedLegOffset)
	}
	// If selected leg is below the viewport, scroll down to show it
	if selectedLegOffset+selectedLegHeight > viewportBottom {
		newOffset := selectedLegOffset + selectedLegHeight - s.viewport.Height
		if newOffset < 0 {
			newOffset = 0
		}
		s.viewport.SetYOffset(newOffset)
	}
}

// displayRouteWithOffsetsAndHeights formats the details of a single journey and tracks leg positions and heights
func (s *routeState) displayRouteWithOffsetsAndHeights(r api.Journey) (string, []int, []int) {
	var doc strings.Builder
	var offsets []int
	var heights []int

	if len(r.Legs) == 0 {
		return "This journey has no legs.", []int{}, []int{}
	}

	origin := r.Legs[0].Origin
	destination := r.Legs[len(r.Legs)-1].Destination

	title := fmt.Sprintf("%s @%s\n\n%s @%s\n\n", origin.DisassembledName, formatTime(s.loc, origin.DepartureTimeEstimated), destination.DisassembledName, formatTime(s.loc, destination.ArrivalTimeEstimated))
	doc.WriteString(title)

	// Count lines in title for offset calculation
	titleLines := strings.Count(title, "\n")
	currentOffset := titleLines

	for idx, leg := range r.Legs {
		offsets = append(offsets, currentOffset)
		legContent := s.formatLeg(leg, idx)
		doc.WriteString(legContent)

		// Calculate actual height of this leg by counting newlines in the rendered content
		legHeight := strings.Count(legContent, "\n")
		heights = append(heights, legHeight)
		currentOffset += legHeight
	}

	return doc.String(), offsets, heights
}

// displayRouteWithOffsets formats the details of a single journey and tracks leg positions
func (s *routeState) displayRouteWithOffsets(r api.Journey) (string, []int) {
	content, offsets, _ := s.displayRouteWithOffsetsAndHeights(r)
	return content, offsets
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
func (s *routeState) formatLeg(l api.Leg, idx int) string {
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

	var showSelectedStr string
	isSelected := idx == s.legSelection
	if isSelected {
		showSelectedStr = " (focused)"
	}

	leg := fmt.Sprintf("%s\n\n> Travel for %dmin%s\n\n%s", originStr, duration, showSelectedStr, destStr)

	return styles.FormatRouteLeg(s.legWidth, transport, isSelected).Render(leg) + "\n"
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

	if len(s.Routes) > 0 && s.paginator.Page < len(s.Routes) {
		s.renderLeg(s.Routes[s.paginator.Page].Legs, s.legSelection)
	}
}

// Update handles messages and updates the state.
func (s *routeState) Update(msg tea.Msg) (AppState, tea.Cmd) {
	var cmds []tea.Cmd
	var paginatorCmd, viewportCmd tea.Cmd

	pageBefore := s.paginator.Page
	legSelectionBefore := s.legSelection

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
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
			s.setViewportContent(s.paginator.Page)
			// After resizing, ensure the selected leg is still visible
			s.scrollToSelectedLeg()
		} else {
			s.viewport.SetContent(lipgloss.NewStyle().Width(s.legWidth).Align(lipgloss.Center).Render("No routes found."))
		}
		return s, nil

	case tea.KeyMsg:
		// Handle leg selection keys first
		switch {
		case key.Matches(msg, legSelectionKeymapDefault.NextLeg):
			if len(s.Routes) > 0 && s.paginator.Page < len(s.Routes) {
				maxLeg := len(s.Routes[s.paginator.Page].Legs) - 1
				if s.legSelection < maxLeg {
					s.legSelection++
				}
			}
		case key.Matches(msg, legSelectionKeymapDefault.PrevLeg):
			if s.legSelection > 0 {
				s.legSelection--
			}
		default:
			// For pagination and viewport scrolling (left/right arrows, page up/down)
			if len(s.Routes) > 0 {
				s.paginator, paginatorCmd = s.paginator.Update(msg)
				s.viewport, viewportCmd = s.viewport.Update(msg)
				cmds = append(cmds, paginatorCmd, viewportCmd)
			}
		}
	default:
		// For all other messages, pass them to the children.
		if len(s.Routes) > 0 {
			s.paginator, paginatorCmd = s.paginator.Update(msg)
			s.viewport, viewportCmd = s.viewport.Update(msg)
			cmds = append(cmds, paginatorCmd, viewportCmd)
		}
	}

	// After any potential update, check if the page has changed.
	if len(s.Routes) > 0 && s.paginator.Page != pageBefore {
		s.legSelection = 0 // Reset leg selection when changing routes
		s.setViewportContent(s.paginator.Page)
		s.viewport.GotoTop()
	} else if len(s.Routes) > 0 && s.legSelection != legSelectionBefore {
		// Leg selection changed, update content and scroll to selected leg
		s.setViewportContent(s.paginator.Page)
		s.scrollToSelectedLeg()
	}

	s.RenderCells(s.root.flexBox)
	return s, tea.Batch(cmds...)
}
