package geo

import (
	"encoding/json"
	"math"
	"testing"
)

// 1x1 degree square around (32.5, -83.5) with a small hole in the middle.
const squareWithHole = `{"type":"Polygon","coordinates":[
	[[-84,32],[-83,32],[-83,33],[-84,33],[-84,32]],
	[[-83.6,32.4],[-83.4,32.4],[-83.4,32.6],[-83.6,32.6],[-83.6,32.4]]
]}`

func decode(t *testing.T, s string) Geometry {
	t.Helper()
	var g Geometry
	if err := json.Unmarshal([]byte(s), &g); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestContains(t *testing.T) {
	g := decode(t, squareWithHole)
	cases := []struct {
		name string
		p    Point
		want bool
	}{
		{"inside", Point{32.2, -83.8}, true},
		{"in hole", Point{32.5, -83.5}, false},
		{"outside", Point{34, -83.5}, false},
	}
	for _, c := range cases {
		if got := g.Contains(c.p); got != c.want {
			t.Errorf("%s: Contains = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestMultiPolygon(t *testing.T) {
	g := decode(t, `{"type":"MultiPolygon","coordinates":[
		[[[-84,32],[-83,32],[-83,33],[-84,33],[-84,32]]],
		[[[-80,30],[-79,30],[-79,31],[-80,31],[-80,30]]]
	]}`)
	if len(g) != 2 || !g.Contains(Point{30.5, -79.5}) {
		t.Fatalf("second polygon not decoded: %+v", g)
	}
}

func TestNullGeometry(t *testing.T) {
	if g := decode(t, `null`); g != nil {
		t.Fatalf("want nil, got %+v", g)
	}
}

func TestNearest(t *testing.T) {
	g := decode(t, squareWithHole)

	if d, _ := g.Nearest(Point{32.2, -83.8}); d != 0 {
		t.Errorf("inside: distance = %v, want 0", d)
	}

	// One degree of latitude north of the top edge is ~69.1 mi.
	d, pt := g.Nearest(Point{34, -83.5})
	if math.Abs(d-69.1) > 0.5 {
		t.Errorf("north: distance = %v, want ~69.1", d)
	}
	if math.Abs(pt.Lat-33) > 1e-9 || math.Abs(pt.Lon+83.5) > 1e-9 {
		t.Errorf("north: nearest = %+v, want (33, -83.5)", pt)
	}

	// From the hole, the nearest edge is the hole's east/west boundary, 0.1 deg
	// of longitude away: 0.1 * cos(32.5) * 69.1 = ~5.83 mi.
	if d, _ := g.Nearest(Point{32.5, -83.5}); math.Abs(d-5.83) > 0.05 {
		t.Errorf("hole: distance = %v, want ~5.83", d)
	}
}

func TestDirection(t *testing.T) {
	home := Point{32.44, -83.75}
	cases := map[string]Point{
		"N":  {33.0, -83.75},
		"E":  {32.44, -83.0},
		"SW": {32.0, -84.25},
		"W":  {32.44, -84.5},
	}
	for want, p := range cases {
		if got := Direction(home, p); got != want {
			t.Errorf("Direction to %+v = %s, want %s", p, got, want)
		}
	}
}
