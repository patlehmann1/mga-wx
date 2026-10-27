# mga-wx

Severe weather for middle Georgia, in your terminal.

- Active NWS warnings, watches, and advisories within 60 miles, with distance and direction from you (or "YOU'RE INSIDE")
- Today's SPC severe risk at your location
- Current conditions and the next 12 hours, drawn as pixel art
- Rings the terminal bell when a new warning covers you or comes within 10 miles

Alerts refresh every minute. The outlook and forecast refresh every 10 minutes.

## Installation

macOS or Linux, with Homebrew:

```bash
brew install patlehmann1/mga-wx/mga-wx
```

Or download a binary from the [latest release](https://github.com/patlehmann1/mga-wx/releases/latest).

## Usage

```bash
mga-wx              # start the dashboard
mga-wx --version
```

| Key | Action |
|---|---|
| `r` | refresh now |
| `t` | cycle theme (dark, synthwave, cobalt2) |
| `q` | quit |

## Location

By default, distances are measured from the center of ZIP 31069 (Perry, GA). For your exact location, create `~/.config/mga-wx/config.toml`:

```toml
label = "Perry, GA"
lat = 32.44
lon = -83.75
theme = "dark"
```

The Census geocoder turns a street address into coordinates:
https://geocoding.geo.census.gov/geocoder/locations/onelineaddress

## Data sources

- Alerts and forecast: [National Weather Service API](https://www.weather.gov/documentation/services-web-api)
- Severe outlook: [Storm Prediction Center](https://www.spc.noaa.gov/products/outlook/)
