package styles

import lg "github.com/charmbracelet/lipgloss"

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
    T6Colour = "" // TODO: Add T6 colour - not present on site
    T7Colour = lg.Color("#6F818E")
    T8Colour = lg.Color("#00954C")
    T9Colour = lg.Color("#D11F2F")

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
)

// flexbox colours
const (
    InactiveColour = lg.ANSIColor(8)
    ActiveColour = lg.ANSIColor(7)
)
