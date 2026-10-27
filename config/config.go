package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Location struct {
	Label string  `toml:"label"`
	Lat   float64 `toml:"lat"`
	Lon   float64 `toml:"lon"`
}

type Config struct {
	Theme     string     `toml:"theme"`
	Locations []Location `toml:"locations"`
}

// Default uses the Census center point for ZIP 31069 (Perry, GA). Precise
// locations belong in the user's config file, never in this repo.
var Default = Config{
	Theme:     "dark",
	Locations: []Location{{Label: "Perry, GA", Lat: 32.440068, Lon: -83.748603}},
}

func Path() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "mga-wx", "config.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "mga-wx", "config.toml"), nil
}

// Load reads the config file at path, falling back to Default for a missing
// file or any field the file leaves out.
func Load(path string) (Config, error) {
	cfg := Config{Theme: Default.Theme}
	_, err := toml.DecodeFile(path, &cfg)
	if errors.Is(err, fs.ErrNotExist) {
		return Default, nil
	}
	if len(cfg.Locations) == 0 {
		cfg.Locations = Default.Locations
	}
	return cfg, err
}
