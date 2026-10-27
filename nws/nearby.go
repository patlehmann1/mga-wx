package nws

import (
	"cmp"
	"slices"
	"strings"

	"github.com/patlehmann1/mga-wx/geo"
)

type Kind int

const (
	Warning Kind = iota
	Watch
	Advisory
)

type NearbyAlert struct {
	Alert
	Kind Kind
	// InZone means the alert covers our county or forecast zone but has no
	// polygon, so there's no distance to report.
	InZone    bool
	Inside    bool
	Miles     float64
	Direction string
}

func KindOf(event string) Kind {
	switch {
	case strings.HasSuffix(event, "Warning") || strings.HasSuffix(event, "Emergency"):
		return Warning
	case strings.HasSuffix(event, "Watch"):
		return Watch
	}
	return Advisory
}

// Nearby keeps alerts that cover home or come within radiusMi of it, most
// urgent first: warnings before watches before advisories, then closest.
func Nearby(alerts []Alert, home geo.Point, zones []string, radiusMi float64) []NearbyAlert {
	var out []NearbyAlert
	for _, a := range alerts {
		n := NearbyAlert{Alert: a, Kind: KindOf(a.Event)}
		switch {
		case a.Geometry != nil:
			var pt geo.Point
			n.Miles, pt = a.Geometry.Nearest(home)
			n.Inside = n.Miles == 0
			if n.Miles > radiusMi {
				continue
			}
			if !n.Inside {
				n.Direction = geo.Direction(home, pt)
			}
		case slices.ContainsFunc(a.UGC, func(z string) bool { return slices.Contains(zones, z) }):
			n.InZone = true
		default:
			continue
		}
		out = append(out, n)
	}
	slices.SortStableFunc(out, func(a, b NearbyAlert) int {
		if a.Kind != b.Kind {
			return int(a.Kind - b.Kind)
		}
		return cmp.Compare(a.Miles, b.Miles)
	})
	return out
}
