package nws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/patlehmann1/mga-wx/geo"
)

const (
	baseURL   = "https://api.weather.gov"
	userAgent = "mga-wx (github.com/patlehmann1/mga-wx)"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// Location is what NWS knows about a point: the zones used to match alerts
// that have no polygon, and where to get its forecast.
type Location struct {
	Zones          []string
	ForecastHourly string
	RadarStation   string
}

type Alert struct {
	ID       string
	Event    string
	Headline string
	AreaDesc string
	Severity string
	Ends     time.Time
	UGC      []string
	Geometry geo.Geometry
}

type Period struct {
	Start       time.Time
	Temperature int
	WindSpeed   string
	WindDir     string
	Short       string
	Icon        string
	IsDaytime   bool
	PrecipPct   int
}

func get(url string, v any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/geo+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned %d", url, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func FetchLocation(p geo.Point) (Location, error) {
	var r struct {
		Properties struct {
			ForecastHourly string `json:"forecastHourly"`
			ForecastZone   string `json:"forecastZone"`
			County         string `json:"county"`
			RadarStation   string `json:"radarStation"`
		} `json:"properties"`
	}
	if err := get(fmt.Sprintf("%s/points/%.4f,%.4f", baseURL, p.Lat, p.Lon), &r); err != nil {
		return Location{}, err
	}
	pr := r.Properties
	return Location{
		Zones:          []string{path.Base(pr.ForecastZone), path.Base(pr.County)},
		ForecastHourly: pr.ForecastHourly,
		RadarStation:   pr.RadarStation,
	}, nil
}

// FetchAlerts returns every active alert in Georgia. Filtering by distance
// happens in Nearby, since a point query would miss storms just outside it.
func FetchAlerts() ([]Alert, error) {
	var r struct {
		Features []struct {
			Geometry   geo.Geometry `json:"geometry"`
			Properties struct {
				ID       string     `json:"id"`
				Event    string     `json:"event"`
				Headline string     `json:"headline"`
				AreaDesc string     `json:"areaDesc"`
				Severity string     `json:"severity"`
				Ends     *time.Time `json:"ends"`
				Expires  time.Time  `json:"expires"`
				Geocode  struct {
					UGC []string `json:"UGC"`
				} `json:"geocode"`
			} `json:"properties"`
		} `json:"features"`
	}
	if err := get(baseURL+"/alerts/active?area=GA", &r); err != nil {
		return nil, err
	}
	alerts := make([]Alert, 0, len(r.Features))
	for _, f := range r.Features {
		p := f.Properties
		ends := p.Expires
		if p.Ends != nil {
			ends = *p.Ends
		}
		alerts = append(alerts, Alert{
			ID:       p.ID,
			Event:    p.Event,
			Headline: p.Headline,
			AreaDesc: p.AreaDesc,
			Severity: p.Severity,
			Ends:     ends,
			UGC:      p.Geocode.UGC,
			Geometry: f.Geometry,
		})
	}
	return alerts, nil
}

func FetchHourly(url string) ([]Period, error) {
	var r struct {
		Properties struct {
			Periods []struct {
				StartTime     time.Time `json:"startTime"`
				Temperature   int       `json:"temperature"`
				WindSpeed     string    `json:"windSpeed"`
				WindDirection string    `json:"windDirection"`
				ShortForecast string    `json:"shortForecast"`
				Icon          string    `json:"icon"`
				IsDaytime     bool      `json:"isDaytime"`
				Precip        struct {
					Value *int `json:"value"`
				} `json:"probabilityOfPrecipitation"`
			} `json:"periods"`
		} `json:"properties"`
	}
	if err := get(url, &r); err != nil {
		return nil, err
	}
	periods := make([]Period, 0, len(r.Properties.Periods))
	for _, p := range r.Properties.Periods {
		pct := 0
		if p.Precip.Value != nil {
			pct = *p.Precip.Value
		}
		periods = append(periods, Period{
			Start:       p.StartTime,
			Temperature: p.Temperature,
			WindSpeed:   p.WindSpeed,
			WindDir:     p.WindDirection,
			Short:       p.ShortForecast,
			Icon:        IconCode(p.Icon),
			IsDaytime:   p.IsDaytime,
			PrecipPct:   pct,
		})
	}
	return periods, nil
}

// IconCode pulls the primary condition code out of an NWS icon URL, e.g.
// ".../icons/land/day/tsra_hi,20/sct?size=small" gives "tsra_hi".
func IconCode(url string) string {
	_, rest, ok := strings.Cut(url, "/land/")
	if !ok {
		return ""
	}
	parts := strings.Split(rest, "/")
	if len(parts) < 2 {
		return ""
	}
	code, _, _ := strings.Cut(parts[1], ",")
	code, _, _ = strings.Cut(code, "?")
	return code
}
