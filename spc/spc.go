package spc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/patlehmann1/mga-wx/geo"
)

const day1URL = "https://www.spc.noaa.gov/products/outlook/day1otlk_cat.nolyr.geojson"

var levels = []string{"TSTM", "MRGL", "SLGT", "ENH", "MDT", "HIGH"}

// Risk is the SPC Day 1 categorical outlook at a point. A zero Risk means
// no thunderstorms are forecast there.
type Risk struct {
	Label string
	Name  string
	Fill  string
}

type outlook struct {
	Features []struct {
		Geometry   geo.Geometry `json:"geometry"`
		Properties struct {
			Label  string `json:"LABEL"`
			Label2 string `json:"LABEL2"`
			Fill   string `json:"fill"`
		} `json:"properties"`
	} `json:"features"`
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

func FetchDay1(p geo.Point) (Risk, error) {
	resp, err := httpClient.Get(day1URL)
	if err != nil {
		return Risk{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Risk{}, fmt.Errorf("SPC outlook returned %d", resp.StatusCode)
	}
	var o outlook
	if err := json.NewDecoder(resp.Body).Decode(&o); err != nil {
		return Risk{}, err
	}
	return highestAt(o, p), nil
}

func highestAt(o outlook, p geo.Point) Risk {
	var best Risk
	for _, f := range o.Features {
		pr := f.Properties
		if slices.Index(levels, pr.Label) > slices.Index(levels, best.Label) && f.Geometry.Contains(p) {
			best = Risk{Label: pr.Label, Name: pr.Label2, Fill: pr.Fill}
		}
	}
	return best
}
