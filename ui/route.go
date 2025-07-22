package ui

import (
	"context"
	"fmt"
	"strings"
        "github.com/mitchellh/go-wordwrap"
	"github.com/76creates/stickers/flexbox"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/isobelmcrae/trip/api"
	"github.com/isobelmcrae/trip/styles"
)

type routeState struct {
    root *RootModel
    Routes []api.Journey
    legWidth int
}

// using only tfnsw for now
func (s *routeState) getAndDisplayRoutes() []api.Journey {
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
    s := &routeState{
        root: root,
        legWidth: width,
    }

    s.Routes = s.getAndDisplayRoutes()

    return s
}

func (s *routeState) displayRoute(r api.Journey) string {
    var doc strings.Builder

    origin := r.Legs[0].Origin
    destination := r.Legs[len(r.Legs) - 1].Destination

    title := fmt.Sprintf("%s\n\n%s\n\n", origin.DisassembledName, destination.DisassembledName)
    doc.WriteString(title)

    for _, leg := range r.Legs {
        doc.WriteString(s.formatLeg(leg))
    }


    return doc.String()
}

func (s *routeState) formatLeg(l api.Leg) string {
    var doc strings.Builder
    
    // will add route highlighting later
    var transport string
    if l.Transportation.DisassembledName == "" {
        transport = "WALK"
    } else {
        transport = l.Transportation.DisassembledName
    }

    originStr := fmt.Sprintf("[%s] %s\n", transport, l.Origin.DisassembledName)
    destStr := fmt.Sprintf("[%s] %s", transport, l.Destination.DisassembledName)
    
    // format strings further if req.
    wrappedOStr := wordwrap.WrapString(originStr, uint(s.legWidth))
    wrappedDStr := wordwrap.WrapString(destStr, uint(s.legWidth))

    leg := fmt.Sprintf("%s\n%s", wrappedOStr, wrappedDStr)
    
    doc.WriteString(styles.FormatRouteLeg(s.legWidth).Render(leg) + "\n")
    
    return doc.String()
}

func (s *routeState) RenderCells(f *flexbox.FlexBox) {
    f.GetRow(0).GetCell(1).SetContent(styles.RouteTitle.Render(s.displayRoute(s.Routes[0])))
    f.GetRow(0).GetCell(0).SetContent("MAP HERE")
}

func (s *routeState) Update(msg tea.Msg) (AppState, tea.Cmd) {
    var cmd tea.Cmd
    
    return s, cmd
}
