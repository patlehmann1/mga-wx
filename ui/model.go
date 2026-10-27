package ui

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/patlehmann1/mga-wx/config"
	"github.com/patlehmann1/mga-wx/geo"
	"github.com/patlehmann1/mga-wx/nws"
	"github.com/patlehmann1/mga-wx/spc"
)

const (
	alertRadiusMi = 60
	// Warnings this close ring the terminal bell even if they don't cover home.
	bellRadiusMi = 10
	alertsEvery  = time.Minute
	outlookEvery = 10 * time.Minute
	blinkEvery   = 500 * time.Millisecond
)

type (
	locationMsg struct {
		loc nws.Location
		err error
	}
	alertsMsg struct {
		alerts []nws.Alert
		err    error
	}
	riskMsg struct {
		risk spc.Risk
		err  error
	}
	hourlyMsg struct {
		periods []nws.Period
		err     error
	}
	blinkMsg      struct{}
	alertsTickMsg struct{}
	slowTickMsg   struct{}
)

type Model struct {
	cfg   config.Config
	home  geo.Point
	theme int

	loc     *nws.Location
	alerts  []nws.NearbyAlert
	risk    *spc.Risk
	periods []nws.Period
	updated time.Time
	errs    map[string]error
	rung    map[string]bool

	cursorOn      bool
	width, height int
}

func New(cfg config.Config) Model {
	return Model{
		cfg:   cfg,
		home:  geo.Point{Lat: cfg.Lat, Lon: cfg.Lon},
		theme: themeIndex(cfg.Theme),
		errs:  map[string]error{},
		rung:  map[string]bool{},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchLocation(),
		m.fetchRisk(),
		tick(blinkEvery, blinkMsg{}),
		tick(alertsEvery, alertsTickMsg{}),
		tick(outlookEvery, slowTickMsg{}),
	)
}

func tick(d time.Duration, msg tea.Msg) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return msg })
}

func (m Model) fetchLocation() tea.Cmd {
	return func() tea.Msg {
		loc, err := nws.FetchLocation(m.home)
		return locationMsg{loc, err}
	}
}

func (m Model) fetchAlerts() tea.Cmd {
	return func() tea.Msg {
		alerts, err := nws.FetchAlerts()
		return alertsMsg{alerts, err}
	}
}

func (m Model) fetchRisk() tea.Cmd {
	return func() tea.Msg {
		risk, err := spc.FetchDay1(m.home)
		return riskMsg{risk, err}
	}
}

func (m Model) fetchHourly() tea.Cmd {
	url := m.loc.ForecastHourly
	return func() tea.Msg {
		periods, err := nws.FetchHourly(url)
		return hourlyMsg{periods, err}
	}
}

// refresh refetches everything, starting over from the location lookup if
// that hasn't succeeded yet.
func (m Model) refresh() tea.Cmd {
	if m.loc == nil {
		return tea.Batch(m.fetchLocation(), m.fetchRisk())
	}
	return tea.Batch(m.fetchAlerts(), m.fetchRisk(), m.fetchHourly())
}

func ringBell() tea.Msg {
	os.Stdout.WriteString("\a")
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "r":
			return m, m.refresh()
		case "t":
			m.theme = (m.theme + 1) % len(themes)
		}

	case blinkMsg:
		m.cursorOn = !m.cursorOn
		return m, tick(blinkEvery, blinkMsg{})

	case alertsTickMsg:
		next := tick(alertsEvery, alertsTickMsg{})
		if m.loc == nil {
			return m, tea.Batch(next, m.fetchLocation())
		}
		return m, tea.Batch(next, m.fetchAlerts())

	case slowTickMsg:
		cmds := []tea.Cmd{tick(outlookEvery, slowTickMsg{}), m.fetchRisk()}
		if m.loc != nil {
			cmds = append(cmds, m.fetchHourly())
		}
		return m, tea.Batch(cmds...)

	case locationMsg:
		m.errs["location"] = msg.err
		if msg.err != nil {
			return m, nil
		}
		m.loc = &msg.loc
		return m, tea.Batch(m.fetchAlerts(), m.fetchHourly())

	case alertsMsg:
		m.errs["alerts"] = msg.err
		if msg.err != nil {
			return m, nil
		}
		m.alerts = nws.Nearby(msg.alerts, m.home, m.loc.Zones, alertRadiusMi)
		m.updated = time.Now()
		return m, m.bellForNewWarnings()

	case riskMsg:
		m.errs["outlook"] = msg.err
		if msg.err == nil {
			m.risk = &msg.risk
		}

	case hourlyMsg:
		m.errs["forecast"] = msg.err
		if msg.err == nil {
			m.periods = msg.periods
		}
	}
	return m, nil
}

func (m Model) bellForNewWarnings() tea.Cmd {
	ring := false
	for _, a := range m.alerts {
		if m.rung[a.ID] {
			continue
		}
		m.rung[a.ID] = true
		if a.Kind == nws.Warning && (a.Inside || a.InZone || a.Miles <= bellRadiusMi) {
			ring = true
		}
	}
	if ring {
		return ringBell
	}
	return nil
}
