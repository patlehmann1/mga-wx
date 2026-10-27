package ui

import "github.com/charmbracelet/lipgloss"

// Theme colors are converted from the HSL tokens in patlehmann1.github.io's
// globals.css so the TUI matches the site.
type Theme struct {
	Name    string
	Fg      lipgloss.Color
	Primary lipgloss.Color
	Muted   lipgloss.Color
	Danger  lipgloss.Color
	Accent  lipgloss.Color
}

var themes = []Theme{
	{Name: "dark", Fg: "#EEEBE8", Primary: "#D77B42", Muted: "#AFA79D", Danger: "#CF3030", Accent: "#E8A04F"},
	{Name: "synthwave", Fg: "#EEE8EE", Primary: "#FF33BB", Muted: "#AC94B8", Danger: "#FF4D4D", Accent: "#00FFFF"},
	{Name: "cobalt2", Fg: "#F3F7FC", Primary: "#FFC800", Muted: "#A2BBCD", Danger: "#FF618B", Accent: "#FF9D00"},
}

func themeIndex(name string) int {
	for i, t := range themes {
		if t.Name == name {
			return i
		}
	}
	return 0
}

// NES-style box: half-block top and bottom with empty corners, like the
// site's .pixel-box stepped corners.
var pixelBorder = lipgloss.Border{
	Top:         "▄",
	Bottom:      "▀",
	Left:        "█",
	Right:       "█",
	TopLeft:     " ",
	TopRight:    " ",
	BottomLeft:  " ",
	BottomRight: " ",
}
