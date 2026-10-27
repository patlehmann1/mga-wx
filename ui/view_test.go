package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/patlehmann1/mga-wx/config"
	"github.com/patlehmann1/mga-wx/nws"
	"github.com/patlehmann1/mga-wx/spc"
)

var twoSites = config.Config{Locations: []config.Location{
	config.Default.Locations[0],
	{Label: "Lake Oconee", Lat: 33.4, Lon: -83.3},
}}

func populated(width int) Model {
	m := New(twoSites)
	m.width, m.height = width, 40
	s := &m.sites[0]
	s.loc = &nws.Location{Zones: []string{"GAZ095", "GAC153"}}
	s.risk = &spc.Risk{Label: "SLGT", Name: "Slight Risk", Fill: "#F6F67F"}
	m.updated = time.Date(2026, 10, 26, 13, 12, 0, 0, time.Local)
	ends := time.Now().Add(45 * time.Minute)
	s.alerts = []nws.NearbyAlert{
		{Alert: nws.Alert{ID: "1", Event: "Tornado Warning", Ends: ends}, Kind: nws.Warning, Miles: 6.2, Direction: "W"},
		{Alert: nws.Alert{ID: "2", Event: "Severe Thunderstorm Watch", Ends: ends}, Kind: nws.Watch, Inside: true},
		{Alert: nws.Alert{ID: "3", Event: "Wind Advisory", Ends: ends}, Kind: nws.Advisory, InZone: true},
	}
	icons := []string{"skc", "sct", "tsra_hi", "tsra", "rain_showers", "ovc", "fog", "bkn", "snow", "few", "sct", "skc", "ovc"}
	start := time.Now().Truncate(time.Hour)
	for i, icon := range icons {
		s.periods = append(s.periods, nws.Period{
			Start:       start.Add(time.Duration(i) * time.Hour),
			Temperature: 72 - i,
			WindSpeed:   "10 mph",
			WindDir:     "SW",
			Short:       "Chance Showers And Thunderstorms Then Mostly Cloudy",
			Icon:        icon,
			IsDaytime:   i < 5,
			PrecipPct:   i * 10,
		})
	}
	return m
}

func TestViewFitsWidth(t *testing.T) {
	for _, width := range []int{60, 89, 90, 120, 160} {
		view := populated(width).View()
		for i, line := range strings.Split(view, "\n") {
			if w := lipgloss.Width(line); w > width {
				t.Errorf("width %d: line %d is %d wide: %q", width, i, w, ansi.Strip(line))
			}
		}
		if testing.Verbose() {
			t.Logf("width %d:\n%s", width, ansi.Strip(view))
		}
	}
}

func TestViewContent(t *testing.T) {
	view := ansi.Strip(populated(120).View())
	for _, want := range []string{
		"> PERRY, GA",
		"LAKE OCONEE",
		"[tab] location",
		"SLGT",
		"!! TORNADO WARNING",
		"6 mi W",
		"YOU'RE INSIDE",
		"YOUR COUNTY",
		"72°F",
		"NEXT 12H",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q", want)
		}
	}
}

func TestViewLoading(t *testing.T) {
	m := New(config.Default)
	m.width = 100
	if strings.Contains(ansi.Strip(m.View()), "[tab]") {
		t.Error("single location shouldn't show the tab key")
	}
	view := ansi.Strip(m.View())
	for _, want := range []string{"[ ... ]", "loading forecast...", "checking alerts..."} {
		if !strings.Contains(view, want) {
			t.Errorf("loading view missing %q", want)
		}
	}
}

func TestConditionOf(t *testing.T) {
	cases := map[string]condition{
		"skc":          sunny,
		"wind_few":     sunny,
		"sct":          partly,
		"bkn":          partly,
		"ovc":          cloudy,
		"rain_showers": rain,
		"tsra_hi":      storm,
		"tornado":      storm,
		"fog":          fog,
		"snow":         snow,
	}
	for code, want := range cases {
		if got := conditionOf(code); got != want {
			t.Errorf("conditionOf(%q) = %v, want %v", code, got, want)
		}
	}
}

func TestSpritesAreRectangular(t *testing.T) {
	check := func(name string, rows []string, w, h int) {
		if len(rows) != h {
			t.Errorf("%s: %d rows, want %d", name, len(rows), h)
		}
		for i, r := range rows {
			if len(r) != w {
				t.Errorf("%s row %d: width %d, want %d", name, i, len(r), w)
			}
		}
	}
	for c, rows := range bigSprites {
		check("big "+string(rune('0'+c)), rows, 16, 10)
	}
	for c, rows := range smallSprites {
		check("small "+string(rune('0'+c)), rows, 6, 4)
	}
	check("big moon", bigMoon, 16, 10)
	check("small moon", smallMoon, 6, 4)
}

func TestBellOnlyForNewNearbyWarnings(t *testing.T) {
	m := populated(100)
	if m.bellForNewWarnings() == nil {
		t.Fatal("want bell for new warning 6 mi away")
	}
	if m.bellForNewWarnings() != nil {
		t.Fatal("bell rang twice for the same warning")
	}

	far := New(config.Default)
	far.sites[0].alerts = []nws.NearbyAlert{{Alert: nws.Alert{ID: "x"}, Kind: nws.Warning, Miles: 30, Direction: "N"}}
	if far.bellForNewWarnings() != nil {
		t.Fatal("bell rang for a warning 30 mi away")
	}
}

func TestTabsSwitchAndFlagWarnings(t *testing.T) {
	m := populated(120)
	m.sites[1].alerts = []nws.NearbyAlert{{Alert: nws.Alert{ID: "lake", Event: "Tornado Warning"}, Kind: nws.Warning, Inside: true}}
	m.sites[0].alerts = m.sites[0].alerts[1:]

	header := ansi.Strip(m.header(themes[0]))
	if !strings.Contains(header, "LAKE OCONEE !!") || strings.Contains(header, "PERRY, GA !!") {
		t.Fatalf("only the lake tab should be flagged: %q", header)
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	view := ansi.Strip(next.View())
	if !strings.Contains(view, "> LAKE OCONEE") || !strings.Contains(view, "YOU'RE INSIDE") {
		t.Fatalf("tab should switch to the lake site:\n%s", view)
	}
	if !strings.Contains(view, "[ ... ]") {
		t.Errorf("lake has no outlook yet; badge should be loading")
	}

	back, _ := next.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if back.(Model).active != 0 {
		t.Errorf("shift+tab should wrap back to the first site")
	}
}

func TestBellPerSite(t *testing.T) {
	m := New(twoSites)
	warning := nws.NearbyAlert{Alert: nws.Alert{ID: "w"}, Kind: nws.Warning, InZone: true}
	m.sites[0].alerts = []nws.NearbyAlert{warning}
	if m.bellForNewWarnings() == nil {
		t.Fatal("want bell for warning at first site")
	}
	m.sites[1].alerts = []nws.NearbyAlert{warning}
	if m.bellForNewWarnings() == nil {
		t.Fatal("same warning newly covering the second site should ring again")
	}
}
