package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type condition int

const (
	sunny condition = iota
	partly
	cloudy
	rain
	storm
	fog
	snow
)

// conditionOf maps an NWS icon code (see nws.IconCode) to a sprite.
func conditionOf(code string) condition {
	code = strings.TrimPrefix(code, "wind_")
	has := func(subs ...string) bool {
		for _, s := range subs {
			if strings.Contains(code, s) {
				return true
			}
		}
		return false
	}
	switch {
	case has("tsra", "tornado", "hurricane", "tropical"):
		return storm
	case has("snow", "sleet", "blizzard"):
		return snow
	case has("rain", "fzra"):
		return rain
	case has("fog", "haze", "smoke", "dust"):
		return fog
	case code == "ovc":
		return cloudy
	case code == "sct" || code == "bkn":
		return partly
	}
	return sunny
}

var pixelColors = map[byte]lipgloss.Color{
	'Y': "#FFD23F", // sun
	'O': "#FF9F1C", // sun rim
	'M': "#F4E9C1", // moon
	'W': "#E8E8E8", // cloud
	'G': "#9AA0A6", // rain cloud
	'D': "#5F6368", // storm cloud
	'B': "#4FA3FF", // rain
	'L': "#FFE45C", // lightning
	'F': "#B0B0B0", // fog
}

// Sprites are pixel grids; each pair of rows becomes one line of half-blocks.
var bigSprites = map[condition][]string{
	sunny: {
		".......YY.......",
		"..Y....YY....Y..",
		"...Y.OOOOOO.Y...",
		"....OOYYYYOO....",
		"YY.OYYYYYYYYO.YY",
		"YY.OYYYYYYYYO.YY",
		"....OOYYYYOO....",
		"...Y.OOOOOO.Y...",
		"..Y....YY....Y..",
		".......YY.......",
	},
	partly: {
		"..Y..Y..........",
		"...OOOO.........",
		"Y.OYYYYO........",
		"..OYYYWWWW......",
		"..OYWWWWWWWW....",
		"...WWWWWWWWWWWW.",
		"..WWWWWWWWWWWWWW",
		"..WWWWWWWWWWWWWW",
		"...GGGGGGGGGGGG.",
		"................",
	},
	cloudy: {
		"................",
		"................",
		"......WWWW......",
		"....WWWWWWW.....",
		"...WWWWWWWWWWW..",
		"..WWWWWWWWWWWWW.",
		".WWWWWWWWWWWWWWW",
		".WWWWWWWWWWWWWWW",
		"..GGGGGGGGGGGGG.",
		"................",
	},
	rain: {
		"......GGGG......",
		"....GGGGGGG.....",
		"...GGGGGGGGGGG..",
		"..GGGGGGGGGGGGG.",
		".GGGGGGGGGGGGGGG",
		"..DDDDDDDDDDDDD.",
		"...B...B...B....",
		"..B...B...B.....",
		".....B...B...B..",
		"....B...B...B...",
	},
	storm: {
		"......DDDD......",
		"....DDDDDDD.....",
		"...DDDDDDDDDDD..",
		"..DDDDDDDDDDDDD.",
		".DDDDDDDDDDDDDDD",
		"..DDDDDLLDDDDDD.",
		"......LL....B...",
		".....LLLL..B....",
		".B.....LL.......",
		"......L.........",
	},
	fog: {
		"................",
		"..FFFFFFFFFFF...",
		"................",
		"....FFFFFFFFFFFF",
		"................",
		".FFFFFFFFFFFF...",
		"................",
		"...FFFFFFFFFFF..",
		"................",
		"................",
	},
	snow: {
		"......GGGG......",
		"....GGGGGGG.....",
		"...GGGGGGGGGGG..",
		"..GGGGGGGGGGGGG.",
		".GGGGGGGGGGGGGGG",
		"..GGGGGGGGGGGGG.",
		"...W....W....W..",
		"................",
		".....W....W.....",
		"................",
	},
}

var bigMoon = []string{
	"......MMMM......",
	"....MMMMM.......",
	"...MMMM.........",
	"..MMMM.......Y..",
	"..MMM...........",
	"..MMM.....Y.....",
	"..MMMM..........",
	"...MMMMM....MM..",
	"....MMMMMMMMMM..",
	"......MMMMMM....",
}

var smallSprites = map[condition][]string{
	sunny:  {"..YY..", ".YYYY.", ".YYYY.", "..YY.."},
	partly: {".YY...", "YYWW..", "YWWWW.", "WWWWWW"},
	cloudy: {"......", "..WW..", ".WWWW.", "WWWWWW"},
	rain:   {".GGGG.", "GGGGGG", ".B..B.", "B..B.."},
	storm:  {".DDDD.", "DDDDDD", "..LL..", ".L...."},
	fog:    {"FFFFF.", "......", ".FFFFF", "......"},
	snow:   {".GGGG.", "GGGGGG", "W..W..", "..W..W"},
}

var smallMoon = []string{"..MMM.", ".MM...", ".MM...", "..MMM."}

var toNight = strings.NewReplacer("Y", "M", "O", "M")

func sprite(set map[condition][]string, moon []string, c condition, day bool) []string {
	switch {
	case day:
		return set[c]
	case c == sunny:
		return moon
	case c == partly:
		night := make([]string, len(set[c]))
		for i, row := range set[c] {
			night[i] = toNight.Replace(row)
		}
		return night
	}
	return set[c]
}

func bigSprite(code string, day bool) string {
	return renderPixels(sprite(bigSprites, bigMoon, conditionOf(code), day))
}

func smallSprite(code string, day bool) string {
	return renderPixels(sprite(smallSprites, smallMoon, conditionOf(code), day))
}

func renderPixels(rows []string) string {
	var b strings.Builder
	for y := 0; y+1 < len(rows); y += 2 {
		if y > 0 {
			b.WriteByte('\n')
		}
		top, bot := rows[y], rows[y+1]
		for x := 0; x < len(top); x++ {
			b.WriteString(halfBlock(top[x], bot[x]))
		}
	}
	return b.String()
}

func halfBlock(top, bot byte) string {
	switch {
	case top == '.' && bot == '.':
		return " "
	case top == bot:
		return lipgloss.NewStyle().Foreground(pixelColors[top]).Render("█")
	case top == '.':
		return lipgloss.NewStyle().Foreground(pixelColors[bot]).Render("▄")
	case bot == '.':
		return lipgloss.NewStyle().Foreground(pixelColors[top]).Render("▀")
	}
	return lipgloss.NewStyle().Foreground(pixelColors[top]).Background(pixelColors[bot]).Render("▀")
}
