package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/patlehmann1/mga-wx/nws"
)

const (
	nowWidth     = 40
	twoColumnMin = 90
	hourWidth    = 7
	maxHours     = 12
	maxAlerts    = 5
)

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	th := themes[m.theme]
	w := m.width

	var middle string
	if w >= twoColumnMin {
		nowBody, alertBody := m.nowBody(th), m.alertsBody(th, w-nowWidth-1)
		h := max(lipgloss.Height(nowBody), lipgloss.Height(alertBody))
		middle = lipgloss.JoinHorizontal(lipgloss.Top,
			box(th, "NOW", nowBody, nowWidth, h),
			" ",
			box(th, "ALERTS", alertBody, w-nowWidth-1, h),
		)
	} else {
		middle = lipgloss.JoinVertical(lipgloss.Left,
			box(th, "NOW", m.nowBody(th), w, 0),
			box(th, "ALERTS", m.alertsBody(th, w), w, 0),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.header(th),
		middle,
		box(th, fmt.Sprintf("NEXT %dH", m.hourCount()), m.hourlyBody(th), w, 0),
		m.footer(th),
	)
}

// box draws an NES-style pixel frame with a "> TITLE" prompt heading, like
// the site's .pixel-box and .text-heading. height 0 sizes to content.
func box(th Theme, title, body string, width, height int) string {
	heading := lipgloss.NewStyle().Foreground(th.Primary).Render("> ") +
		lipgloss.NewStyle().Foreground(th.Fg).Bold(true).Render(title)
	style := lipgloss.NewStyle().
		Border(pixelBorder).
		BorderForeground(th.Muted).
		Padding(0, 1).
		Width(width - 2)
	if height > 0 {
		style = style.Height(height + 1)
	}
	return style.Render(heading + "\n" + body)
}

func innerWidth(boxWidth int) int {
	return boxWidth - 4
}

func (m Model) header(th Theme) string {
	cursor := " "
	if m.cursorOn {
		cursor = lipgloss.NewStyle().Foreground(th.Primary).Render("█")
	}
	left := lipgloss.NewStyle().Foreground(th.Primary).Render("> ") +
		lipgloss.NewStyle().Foreground(th.Fg).Bold(true).Render(strings.ToUpper(m.cfg.Label)+" WX") +
		" " + cursor

	label := lipgloss.NewStyle().Foreground(th.Muted).Render("SPC DAY 1 ")
	var badge string
	switch {
	case m.risk == nil:
		badge = lipgloss.NewStyle().Foreground(th.Muted).Render("[ ... ]")
	case m.risk.Label == "":
		badge = lipgloss.NewStyle().Foreground(th.Muted).Render("[ NONE ]")
	default:
		badge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1A140F")).
			Background(lipgloss.Color(m.risk.Fill)).
			Bold(true).
			Render(" " + m.risk.Label + " ")
	}
	right := label + badge

	gap := max(1, innerWidth(m.width)-lipgloss.Width(left)-lipgloss.Width(right))
	return lipgloss.NewStyle().
		Border(pixelBorder).
		BorderForeground(th.Primary).
		Padding(0, 1).
		Width(m.width - 2).
		Render(left + strings.Repeat(" ", gap) + right)
}

func (m Model) nowBody(th Theme) string {
	muted := lipgloss.NewStyle().Foreground(th.Muted)
	if len(m.periods) == 0 {
		return muted.Render("loading forecast...")
	}
	p := m.periods[0]

	temp := lipgloss.NewStyle().Foreground(th.Fg).Bold(true).Render(fmt.Sprintf("%d°F", p.Temperature))
	details := lipgloss.JoinVertical(lipgloss.Left,
		temp,
		"",
		muted.Render("wind "+p.WindDir+" "+p.WindSpeed),
		muted.Render(fmt.Sprintf("rain %d%%", p.PrecipPct)),
	)
	art := lipgloss.JoinHorizontal(lipgloss.Top, bigSprite(p.Icon, p.IsDaytime), "   ", details)
	short := ansi.Truncate(p.Short, innerWidth(nowWidth), "…")
	return art + "\n\n" + lipgloss.NewStyle().Foreground(th.Accent).Render(short)
}

func (m Model) alertsBody(th Theme, boxWidth int) string {
	muted := lipgloss.NewStyle().Foreground(th.Muted)
	if m.loc == nil || m.updated.IsZero() {
		return muted.Render("checking alerts...")
	}
	if len(m.alerts) == 0 {
		return lipgloss.NewStyle().Foreground(th.Fg).Render("All clear.") + "\n" +
			muted.Render(fmt.Sprintf("No alerts within %d mi.", alertRadiusMi))
	}

	width := innerWidth(boxWidth)
	var lines []string
	for i, a := range m.alerts {
		if i == maxAlerts {
			lines = append(lines, muted.Render(fmt.Sprintf("+%d more", len(m.alerts)-maxAlerts)))
			break
		}
		marker, color := "·  ", th.Muted
		switch a.Kind {
		case nws.Warning:
			marker, color = "!! ", th.Danger
		case nws.Watch:
			marker, color = "!  ", th.Primary
		}
		title := lipgloss.NewStyle().Foreground(color).Bold(a.Kind == nws.Warning).
			Render(ansi.Truncate(marker+strings.ToUpper(a.Event), width, "…"))

		where := lipgloss.NewStyle().Foreground(color).Bold(true).Render(whereText(a))
		until := muted.Render(" · until " + a.Ends.Local().Format("Mon 3:04 PM"))
		lines = append(lines, title, ansi.Truncate("   "+where+until, width, "…"))
	}
	return strings.Join(lines, "\n")
}

func whereText(a nws.NearbyAlert) string {
	switch {
	case a.Inside:
		return "YOU'RE INSIDE"
	case a.InZone:
		return "YOUR COUNTY"
	case a.Miles < 1:
		return "<1 mi " + a.Direction
	}
	return fmt.Sprintf("%.0f mi %s", a.Miles, a.Direction)
}

func (m Model) hourCount() int {
	return max(1, min(maxHours, innerWidth(m.width)/hourWidth))
}

func (m Model) hourlyBody(th Theme) string {
	muted := lipgloss.NewStyle().Foreground(th.Muted)
	cutoff := time.Now().Add(-time.Hour)
	var cols []string
	for _, p := range m.periods {
		if len(cols) == m.hourCount() {
			break
		}
		if p.Start.Before(cutoff) {
			continue
		}
		precip := muted.Render(fmt.Sprintf("%d%%", p.PrecipPct))
		if p.PrecipPct >= 30 {
			precip = lipgloss.NewStyle().Foreground(pixelColors['B']).Render(fmt.Sprintf("%d%%", p.PrecipPct))
		}
		col := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Foreground(th.Primary).Render(hourLabel(p.Start)),
			smallSprite(p.Icon, p.IsDaytime),
			lipgloss.NewStyle().Foreground(th.Fg).Render(fmt.Sprintf("%d°", p.Temperature)),
			precip,
		)
		cols = append(cols, lipgloss.NewStyle().Width(hourWidth).Render(col))
	}
	if len(cols) == 0 {
		return muted.Render("loading forecast...")
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

// hourLabel formats 3:00 PM as "3p" and midnight as "12a".
func hourLabel(t time.Time) string {
	return strings.TrimSuffix(t.Local().Format("3pm"), "m")
}

func (m Model) footer(th Theme) string {
	muted := lipgloss.NewStyle().Foreground(th.Muted)
	key := lipgloss.NewStyle().Foreground(th.Primary)
	left := " " + key.Render("[r]") + muted.Render(" refresh  ") +
		key.Render("[t]") + muted.Render(" theme: "+themes[m.theme].Name+"  ") +
		key.Render("[q]") + muted.Render(" quit")

	var right string
	for _, src := range []string{"location", "alerts", "outlook", "forecast"} {
		if err := m.errs[src]; err != nil {
			right = lipgloss.NewStyle().Foreground(th.Danger).Render(src + " fetch failed ")
			break
		}
	}
	if right == "" && !m.updated.IsZero() {
		right = muted.Render("updated " + m.updated.Format("3:04 PM") + " ")
	}

	gap := max(1, m.width-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", gap) + right
}
