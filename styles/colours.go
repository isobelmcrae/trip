package styles

import lg "github.com/charmbracelet/lipgloss"
import "strings"

// TODO: some way to map colours to line name/id/mode something
// that corresponds to the output given in the `/trip` endpoint
// + add T6 colour

// colours for each transport line
// source: https://opendata.transport.nsw.gov.au/developers/resources
const (
    // metro
    MetroColour = "#168388"

    // sydney trains
    T1Colour = "#F99D1C"
    T2Colour = "#0098CD"
    T3Colour = "#F37021"
    T4Colour = "#005AA"
    T5Colour = "#C4258F"
    T6Colour = "#7F3D1B"
    T7Colour = "#6F818E"
    T8Colour = "#00954C"
    T9Colour = "#D11F2F"

    BusColour = "#009ED7"

    // intercity trains
    BlueMountainsColour = "#F99D1C"
    CCNewcastleColour = "#D11F2F"
    HunterColour = "#833134"
    SouthCoastColour = "#005AA3"
    SouthernHighlandsColour = "#00954C"

    // regional trains and coaches network
    TrainsColour = "#F6891F"
    CoachesColour = "#732A82"

    // ferries
    F1Colour = "#00774B"
    F2Colour = "#144734"
    F3Colour = "#648C3C"
    F4Colour = "#BFD730"
    F5Colour = "#286142"
    F6Colour = "#00AB51"
    F7Colour = "#00B189"
    F8Colour = "#55622B"
    F9Colour = "#65B32E"
    F10Colour = "#5AB031" // colour spec subject to change
    StocktonColour = "#5AB031"

    // sydney light rail
    L1Colour = "#BE1622"
    L2Colour = "#DD1E25"
    L3Colour = "#781140"
    NLRColour = "#EE343F"

    WalkColour = "#4d4d4d"
)

// flexbox colours
const (
    InactiveColour = lg.ANSIColor(8)
    ActiveColour = lg.ANSIColor(7)
)

var LineColours = map[string]string{
    // Metro
    "Metro": MetroColour,

    "Bus": BusColour,

    // Sydney Trains
    "T1": T1Colour,
    "T2": T2Colour,
    "T3": T3Colour,
    "T4": T4Colour,
    "T5": T5Colour,
    "T6": T6Colour,
    "T7": T7Colour,
    "T8": T8Colour,
    "T9": T9Colour,

    // Intercity
    "BlueMountains": BlueMountainsColour,
    "CCNewcastle":   CCNewcastleColour,
    "Hunter":        HunterColour,
    "SouthCoast":    SouthCoastColour,
    "SouthernHighlands": SouthernHighlandsColour,

    // Regional
    "Trains":  TrainsColour,
    "Coaches": CoachesColour,

    // Ferries
    "F1":      F1Colour,
    "F2":      F2Colour,
    "F3":      F3Colour,
    "F4":      F4Colour,
    "F5":      F5Colour,
    "F6":      F6Colour,
    "F7":      F7Colour,
    "F8":      F8Colour,
    "F9":      F9Colour,
    "F10":     F10Colour,
    "Stockton": StocktonColour,

    // Light Rail
    "L1":  L1Colour,
    "L2":  L2Colour,
    "L3":  L3Colour,
    "NLR": NLRColour,

    "WALK": WalkColour,
}

func LgColourForLine(line string) lg.Color {
    line = strings.ToUpper(line)
    var colour string

    switch {
    // Metro lines like M1, M2, M3
    case strings.HasPrefix(line, "M"):
        colour = MetroColour

    // Ferries (F1-F10)
    case strings.HasPrefix(line, "F"):
        if c, ok := LineColours[line]; ok {
            colour = c
        }

    // Light rail
    case strings.HasPrefix(line, "L"):
        if c, ok := LineColours[line]; ok {
            colour = c
        }

    // Trains (T1-T9)
    case strings.HasPrefix(line, "T"):
        if c, ok := LineColours[line]; ok {
            colour = c
        }

    // Explicit map fallback (covers intercity, regional, named lines)
    default:
        if c, ok := LineColours[line]; ok {
            colour = c
        } else if line == "WALK" {
            colour = LineColours["WALK"]
        } else {
            colour = LineColours["Bus"]
        }
    }

    return lg.Color(colour)
}


func HexColourForLine(line string) string {
    line = strings.ToUpper(line)
    var colour string

    switch {
    // Metro lines like M1, M2, M3
    case strings.HasPrefix(line, "M"):
        colour = MetroColour

    // Ferries (F1-F10)
    case strings.HasPrefix(line, "F"):
        if c, ok := LineColours[line]; ok {
            colour = c
        }

    // Light rail
    case strings.HasPrefix(line, "L"):
        if c, ok := LineColours[line]; ok {
            colour = c
        }

    // Trains (T1-T9)
    case strings.HasPrefix(line, "T"):
        if c, ok := LineColours[line]; ok {
            colour = c
        }

    // Explicit map fallback (covers intercity, regional, named lines)
    default:
        if c, ok := LineColours[line]; ok {
            colour = c
        } else if line == "WALK" {
            colour = LineColours["WALK"]
        } else {
            colour = LineColours["Bus"]
        }
    }

    return colour
}


