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
    MetroColour = lg.Color("#168388")

    // sydney trains
    T1Colour = lg.Color("#F99D1C")
    T2Colour = lg.Color("#0098CD")
    T3Colour = lg.Color("#F37021")
    T4Colour = lg.Color("#005AA")
    T5Colour = lg.Color("#C4258F")
    T6Colour = lg.Color("#7F3D1B")
    T7Colour = lg.Color("#6F818E")
    T8Colour = lg.Color("#00954C")
    T9Colour = lg.Color("#D11F2F")

    BusColour = lg.Color("#009ED7")

    // intercity trains
    BlueMountainsColour = lg.Color("#F99D1C")
    CCNewcastleColour = lg.Color("#D11F2F")
    HunterColour = lg.Color("#833134")
    SouthCoastColour = lg.Color("#005AA3")
    SouthernHighlandsColour = lg.Color("#00954C")

    // regional trains and coaches network
    TrainsColour = lg.Color("#F6891F")
    CoachesColour = lg.Color("#732A82")

    // ferries
    F1Colour = lg.Color("#00774B")
    F2Colour = lg.Color("#144734")
    F3Colour = lg.Color("#648C3C")
    F4Colour = lg.Color("#BFD730")
    F5Colour = lg.Color("#286142")
    F6Colour = lg.Color("#00AB51")
    F7Colour = lg.Color("#00B189")
    F8Colour = lg.Color("#55622B")
    F9Colour = lg.Color("#65B32E")
    F10Colour = lg.Color("#5AB031") // colour spec subject to change
    StocktonColour = lg.Color("#5AB031")

    // sydney light rail
    L1Colour = lg.Color("#BE1622")
    L2Colour = lg.Color("#DD1E25")
    L3Colour = lg.Color("#781140")
    NLRColour = lg.Color("#EE343F")

    WalkColour = lg.Color("#4d4d4d")
)

// flexbox colours
const (
    InactiveColour = lg.ANSIColor(8)
    ActiveColour = lg.ANSIColor(7)
)

var LineColours = map[string]lg.Color{
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

func ColourForLine(line string) lg.Color {
    line = strings.ToUpper(line)

    switch {
    // Metro lines like M1, M2, M3
    case strings.HasPrefix(line, "M"):
        return MetroColour

    // Ferries (F1-F10)
    case strings.HasPrefix(line, "F"):
        if c, ok := LineColours[line]; ok {
            return c
        }

    // Light rail
    case strings.HasPrefix(line, "L"):
        if c, ok := LineColours[line]; ok {
            return c
        }

    // Trains (T1-T9)
    case strings.HasPrefix(line, "T"):
        if c, ok := LineColours[line]; ok {
            return c
        }

    // Explicit map fallback (covers intercity, regional, named lines)
    default:
        if c, ok := LineColours[line]; ok {
            return c
        } else if line == "WALK" {
            return LineColours["WALK"]
        } else {
            return LineColours["Bus"]
        }
    }

    return lg.Color("") // default: no colour
}

