package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Label string  `toml:"label"`
	Lat   float64 `toml:"lat"`
	Lon   float64 `toml:"lon"`
	Theme string  `toml:"theme"`
}

// Default is the Census center point for ZIP 31069 (Perry, GA). Precise
// locations belong in the user's config file, never in this repo.
var Default = Config{
	Label: "Perry, GA",
	Lat:   32.440068,
	Lon:   -83.748603,
	Theme: "dark",
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
	cfg := Default
	_, err := toml.DecodeFile(path, &cfg)
	if errors.Is(err, fs.ErrNotExist) {
		return Default, nil
	}
	return cfg, err
}
