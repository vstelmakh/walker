// Package config holds the user settings.
package config

type Config struct {
	// ShowHidden lists the directories whose names start with a dot.
	ShowHidden bool
}

// Default returns the built-in settings.
func Default() Config {
	return Config{
		ShowHidden: false,
	}
}
