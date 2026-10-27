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
		site int
		loc  nws.Location
		err  error
	}
	alertsMsg struct {
		alerts []nws.Alert
		err    error
	}
	outlookMsg struct {
		outlook spc.Outlook
		err     error
	}
	hourlyMsg struct {
		site    int
		periods []nws.Period
		err     error
	}
	blinkMsg      struct{}
	alertsTickMsg struct{}
	slowTickMsg   struct{}
)

type site struct {
	label   string
	home    geo.Point
	loc     *nws.Location
	alerts  []nws.NearbyAlert
	risk    *spc.Risk
	periods []nws.Period
}

type Model struct {
	sites  []site
	active int
	theme  int

	allAlerts []nws.Alert
	updated   time.Time
	errs      map[string]error
	rung      map[string]bool

	cursorOn      bool
	width, height int
}

func New(cfg config.Config) Model {
	m := Model{
		theme: themeIndex(cfg.Theme),
		errs:  map[string]error{},
		rung:  map[string]bool{},
	}
	for _, l := range cfg.Locations {
		m.sites = append(m.sites, site{label: l.Label, home: geo.Point{Lat: l.Lat, Lon: l.Lon}})
	}
	return m
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		fetchAlerts,
		fetchOutlook,
		tick(blinkEvery, blinkMsg{}),
		tick(alertsEvery, alertsTickMsg{}),
		tick(outlookEvery, slowTickMsg{}),
	}
	for i := range m.sites {
		cmds = append(cmds, m.fetchLocation(i))
	}
	return tea.Batch(cmds...)
}

func tick(d time.Duration, msg tea.Msg) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return msg })
}

func (m Model) fetchLocation(i int) tea.Cmd {
	home := m.sites[i].home
	return func() tea.Msg {
		loc, err := nws.FetchLocation(home)
		return locationMsg{i, loc, err}
	}
}

func fetchAlerts() tea.Msg {
	alerts, err := nws.FetchAlerts()
	return alertsMsg{alerts, err}
}

func fetchOutlook() tea.Msg {
	outlook, err := spc.FetchDay1()
	return outlookMsg{outlook, err}
}

func (m Model) fetchHourly(i int) tea.Cmd {
	url := m.sites[i].loc.ForecastHourly
	return func() tea.Msg {
		periods, err := nws.FetchHourly(url)
		return hourlyMsg{i, periods, err}
	}
}

// perSite fetches each site's forecast, or its location first if that
// lookup hasn't succeeded yet.
func (m Model) perSite(withHourly bool) []tea.Cmd {
	var cmds []tea.Cmd
	for i, s := range m.sites {
		switch {
		case s.loc == nil:
			cmds = append(cmds, m.fetchLocation(i))
		case withHourly:
			cmds = append(cmds, m.fetchHourly(i))
		}
	}
	return cmds
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
			return m, tea.Batch(append(m.perSite(true), fetchAlerts, fetchOutlook)...)
		case "t":
			m.theme = (m.theme + 1) % len(themes)
		case "tab", "right":
			m.active = (m.active + 1) % len(m.sites)
		case "shift+tab", "left":
			m.active = (m.active + len(m.sites) - 1) % len(m.sites)
		}

	case blinkMsg:
		m.cursorOn = !m.cursorOn
		return m, tick(blinkEvery, blinkMsg{})

	case alertsTickMsg:
		cmds := append(m.perSite(false), tick(alertsEvery, alertsTickMsg{}), fetchAlerts)
		return m, tea.Batch(cmds...)

	case slowTickMsg:
		cmds := append(m.perSite(true), tick(outlookEvery, slowTickMsg{}), fetchOutlook)
		return m, tea.Batch(cmds...)

	case locationMsg:
		s := &m.sites[msg.site]
		m.errs[s.label+" location"] = msg.err
		if msg.err != nil {
			return m, nil
		}
		s.loc = &msg.loc
		m.matchAlerts()
		return m, tea.Batch(m.fetchHourly(msg.site), m.bellForNewWarnings())

	case alertsMsg:
		m.errs["alerts"] = msg.err
		if msg.err != nil {
			return m, nil
		}
		m.allAlerts = msg.alerts
		m.updated = time.Now()
		m.matchAlerts()
		return m, m.bellForNewWarnings()

	case outlookMsg:
		m.errs["outlook"] = msg.err
		if msg.err == nil {
			for i := range m.sites {
				risk := msg.outlook.RiskAt(m.sites[i].home)
				m.sites[i].risk = &risk
			}
		}

	case hourlyMsg:
		s := &m.sites[msg.site]
		m.errs[s.label+" forecast"] = msg.err
		if msg.err == nil {
			s.periods = msg.periods
		}
	}
	return m, nil
}

// matchAlerts filters the statewide alert list for each site. Zone matching
// needs the site's NWS location, so sites still waiting on it only get
// polygon matches.
func (m Model) matchAlerts() {
	for i := range m.sites {
		s := &m.sites[i]
		var zones []string
		if s.loc != nil {
			zones = s.loc.Zones
		}
		s.alerts = nws.Nearby(m.allAlerts, s.home, zones, alertRadiusMi)
	}
}

func affects(a nws.NearbyAlert) bool {
	return a.Kind == nws.Warning && (a.Inside || a.InZone || a.Miles <= bellRadiusMi)
}

func (m Model) bellForNewWarnings() tea.Cmd {
	ring := false
	for _, s := range m.sites {
		for _, a := range s.alerts {
			key := s.label + "|" + a.ID
			if m.rung[key] {
				continue
			}
			m.rung[key] = true
			ring = ring || affects(a)
		}
	}
	if ring {
		return ringBell
	}
	return nil
}
