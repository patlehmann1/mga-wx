package geo

import (
	"encoding/json"
	"fmt"
	"math"
)

const earthRadiusMi = 3958.8

type Point struct {
	Lat, Lon float64
}

type Ring []Point

// Polygon is an outer ring followed by any hole rings.
type Polygon []Ring

// Geometry decodes a GeoJSON Polygon or MultiPolygon into a list of polygons.
type Geometry []Polygon

func (g *Geometry) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	var raw struct {
		Type        string          `json:"type"`
		Coordinates json.RawMessage `json:"coordinates"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	switch raw.Type {
	case "Polygon":
		var c [][][2]float64
		if err := json.Unmarshal(raw.Coordinates, &c); err != nil {
			return err
		}
		*g = Geometry{toPolygon(c)}
	case "MultiPolygon":
		var c [][][][2]float64
		if err := json.Unmarshal(raw.Coordinates, &c); err != nil {
			return err
		}
		for _, p := range c {
			*g = append(*g, toPolygon(p))
		}
	default:
		return fmt.Errorf("unsupported geometry type %q", raw.Type)
	}
	return nil
}

func toPolygon(c [][][2]float64) Polygon {
	poly := make(Polygon, len(c))
	for i, ring := range c {
		poly[i] = make(Ring, len(ring))
		for j, xy := range ring {
			poly[i][j] = Point{Lat: xy[1], Lon: xy[0]}
		}
	}
	return poly
}

// Contains uses even-odd ray casting across all rings, so holes are excluded.
func (g Geometry) Contains(p Point) bool {
	for _, poly := range g {
		inside := false
		for _, ring := range poly {
			for i, j := 0, len(ring)-1; i < len(ring); j, i = i, i+1 {
				a, b := ring[i], ring[j]
				if (a.Lat > p.Lat) != (b.Lat > p.Lat) &&
					p.Lon < (b.Lon-a.Lon)*(p.Lat-a.Lat)/(b.Lat-a.Lat)+a.Lon {
					inside = !inside
				}
			}
		}
		if inside {
			return true
		}
	}
	return false
}

// Nearest returns the distance in miles from p to the closest edge of g and
// that closest point. It returns 0 and p when p is inside g.
func (g Geometry) Nearest(p Point) (float64, Point) {
	if g.Contains(p) {
		return 0, p
	}
	// Flat projection around p is accurate to well under 1% at the ~100 mi
	// distances this app cares about.
	kx := math.Cos(p.Lat*math.Pi/180) * earthRadiusMi * math.Pi / 180
	ky := earthRadiusMi * math.Pi / 180
	best := math.Inf(1)
	var bestPt Point
	for _, poly := range g {
		for _, ring := range poly {
			for i := 1; i < len(ring); i++ {
				ax, ay := (ring[i-1].Lon-p.Lon)*kx, (ring[i-1].Lat-p.Lat)*ky
				bx, by := (ring[i].Lon-p.Lon)*kx, (ring[i].Lat-p.Lat)*ky
				dx, dy := bx-ax, by-ay
				t := 0.0
				if l := dx*dx + dy*dy; l > 0 {
					t = math.Max(0, math.Min(1, -(ax*dx+ay*dy)/l))
				}
				x, y := ax+t*dx, ay+t*dy
				if d := math.Hypot(x, y); d < best {
					best = d
					bestPt = Point{Lat: p.Lat + y/ky, Lon: p.Lon + x/kx}
				}
			}
		}
	}
	return best, bestPt
}

var compass = [...]string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}

// Direction is the 8-point compass direction from a to b.
func Direction(a, b Point) string {
	la1, la2 := a.Lat*math.Pi/180, b.Lat*math.Pi/180
	dl := (b.Lon - a.Lon) * math.Pi / 180
	y := math.Sin(dl) * math.Cos(la2)
	x := math.Cos(la1)*math.Sin(la2) - math.Sin(la1)*math.Cos(la2)*math.Cos(dl)
	deg := math.Mod(math.Atan2(y, x)*180/math.Pi+360, 360)
	return compass[int(math.Round(deg/45))%8]
}
