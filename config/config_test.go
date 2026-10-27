package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func write(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadMissingFileUsesDefault(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg, Default) {
		t.Fatalf("got %+v, want Default", cfg)
	}
}

func TestLoadLocations(t *testing.T) {
	cfg, err := Load(write(t, `
[[locations]]
label = "Home"
lat = 32.5
lon = -83.7

[[locations]]
label = "Lake"
lat = 33.4
lon = -83.2
`))
	if err != nil {
		t.Fatal(err)
	}
	want := Config{
		Theme: "dark",
		Locations: []Location{
			{Label: "Home", Lat: 32.5, Lon: -83.7},
			{Label: "Lake", Lat: 33.4, Lon: -83.2},
		},
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Fatalf("got %+v, want %+v", cfg, want)
	}
}

func TestLoadNoLocationsUsesDefaultLocation(t *testing.T) {
	cfg, err := Load(write(t, `theme = "cobalt2"`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Theme != "cobalt2" || !reflect.DeepEqual(cfg.Locations, Default.Locations) {
		t.Fatalf("got %+v", cfg)
	}
}

func TestLoadBadFile(t *testing.T) {
	if _, err := Load(write(t, "theme = nope")); err == nil {
		t.Fatal("want error for malformed TOML")
	}
}
