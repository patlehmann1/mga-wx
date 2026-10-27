package nws

import (
	"encoding/json"
	"testing"

	"github.com/patlehmann1/mga-wx/geo"
)

var home = geo.Point{Lat: 32.44, Lon: -83.75}

func square(t *testing.T, minLon, minLat, maxLon, maxLat float64) geo.Geometry {
	t.Helper()
	b, _ := json.Marshal(map[string]any{
		"type": "Polygon",
		"coordinates": [][][2]float64{{
			{minLon, minLat}, {maxLon, minLat}, {maxLon, maxLat}, {minLon, maxLat}, {minLon, minLat},
		}},
	})
	var g geo.Geometry
	if err := json.Unmarshal(b, &g); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestIconCode(t *testing.T) {
	cases := map[string]string{
		"https://api.weather.gov/icons/land/day/tsra_hi,20/sct?size=small":  "tsra_hi",
		"https://api.weather.gov/icons/land/night/sct?size=small":           "sct",
		"https://api.weather.gov/icons/land/day/rain_showers,60?size=small": "rain_showers",
		"": "",
	}
	for url, want := range cases {
		if got := IconCode(url); got != want {
			t.Errorf("IconCode(%q) = %q, want %q", url, got, want)
		}
	}
}

func TestKindOf(t *testing.T) {
	cases := map[string]Kind{
		"Tornado Warning":           Warning,
		"Flash Flood Emergency":     Warning,
		"Severe Thunderstorm Watch": Watch,
		"Wind Advisory":             Advisory,
		"Special Weather Statement": Advisory,
	}
	for event, want := range cases {
		if got := KindOf(event); got != want {
			t.Errorf("KindOf(%q) = %v, want %v", event, got, want)
		}
	}
}

func TestNearby(t *testing.T) {
	alerts := []Alert{
		{ID: "far", Event: "Tornado Warning", Geometry: square(t, -80, 31, -79.9, 31.1)},
		{ID: "advisory-inside", Event: "Wind Advisory", Geometry: square(t, -84, 32, -83, 33)},
		{ID: "watch-zone", Event: "Tornado Watch", UGC: []string{"GAC153"}},
		{ID: "other-zone", Event: "Tornado Watch", UGC: []string{"GAC051"}},
		// Box whose east edge is 0.2 deg of longitude west of home: ~11.7 mi W.
		{ID: "warning-west", Event: "Tornado Warning", Geometry: square(t, -84.2, 32.3, -83.95, 32.6)},
	}

	got := Nearby(alerts, home, []string{"GAZ095", "GAC153"}, 60)

	ids := make([]string, len(got))
	for i, n := range got {
		ids[i] = n.ID
	}
	want := []string{"warning-west", "watch-zone", "advisory-inside"}
	if len(ids) != len(want) {
		t.Fatalf("ids = %v, want %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("ids = %v, want %v", ids, want)
		}
	}

	w := got[0]
	if w.Inside || w.Direction != "W" || w.Miles < 11 || w.Miles > 12.5 {
		t.Errorf("warning-west = %+v, want ~11.7 mi W, not inside", w)
	}
	if !got[1].InZone {
		t.Errorf("watch-zone should be InZone")
	}
	if !got[2].Inside || got[2].Direction != "" {
		t.Errorf("advisory-inside = %+v, want Inside with no direction", got[2])
	}
}
