package spc

import (
	"encoding/json"
	"testing"

	"github.com/patlehmann1/mga-wx/geo"
)

// Nested boxes centered on home: TSTM is widest, SLGT is the core. ENH sits
// off to the east and doesn't cover home. Features are deliberately out of
// order to show the highest level wins regardless of position.
const fixture = `{"features":[
	{"properties":{"LABEL":"SLGT","LABEL2":"Slight Risk","fill":"#F6F67F"},
	 "geometry":{"type":"Polygon","coordinates":[[[-84,32],[-83,32],[-83,33],[-84,33],[-84,32]]]}},
	{"properties":{"LABEL":"TSTM","LABEL2":"General Thunderstorms Risk","fill":"#C1E9C1"},
	 "geometry":{"type":"MultiPolygon","coordinates":[[[[-86,30],[-81,30],[-81,35],[-86,35],[-86,30]]]]}},
	{"properties":{"LABEL":"ENH","LABEL2":"Enhanced Risk","fill":"#E6C27F"},
	 "geometry":{"type":"Polygon","coordinates":[[[-82,32],[-81,32],[-81,33],[-82,33],[-82,32]]]}}
]}`

func TestRiskAt(t *testing.T) {
	var o Outlook
	if err := json.Unmarshal([]byte(fixture), &o); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		p    geo.Point
		want string
	}{
		{"core", geo.Point{Lat: 32.44, Lon: -83.75}, "SLGT"},
		{"east", geo.Point{Lat: 32.5, Lon: -81.5}, "ENH"},
		{"edge", geo.Point{Lat: 34, Lon: -85}, "TSTM"},
		{"none", geo.Point{Lat: 40, Lon: -90}, ""},
	}
	for _, c := range cases {
		if got := o.RiskAt(c.p); got.Label != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got.Label, c.want)
		}
	}
	if r := o.RiskAt(cases[0].p); r.Name != "Slight Risk" || r.Fill != "#F6F67F" {
		t.Errorf("core risk = %+v, want name and fill carried through", r)
	}
}
