package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/vstelmakh/walker/internal/cli"
	"github.com/vstelmakh/walker/internal/config"
	"github.com/vstelmakh/walker/internal/dirs"
	"github.com/vstelmakh/walker/internal/nav"
	"github.com/vstelmakh/walker/internal/search"
	"github.com/vstelmakh/walker/internal/ui"
)

// version is set at build time, for example: -ldflags "-X main.version=1.2.3".
var version = "dev"

var errCancelled = errors.New("cancelled")

func main() {
	if err := run(os.Args[1:]); err != nil {
		if !errors.Is(err, errCancelled) {
			fmt.Fprintln(os.Stderr, "walker:", err)
		}
		os.Exit(1)
	}
}

// run draws the UI on stderr, so stdout carries only the path, for example: cd "$(walker)".
func run(args []string) error {
	opts, cfg, err := cli.Parse(args, config.Default())
	if err != nil {
		return err
	}

	// The help and the version go to stderr, because stdout is for the path. Otherwise, a shell function like
	// cd "$(walker "$@")" would try to change to the help text.
	if text := opts.Info(version); text != "" {
		fmt.Fprint(os.Stderr, text)
		return nil
	}

	start, err := startDir(opts.Dir)
	if err != nil {
		return err
	}

	navigator, err := nav.New(dirs.Lister{}, search.Matcher{}, start, cfg.ShowHidden)
	if err != nil {
		return err
	}

	// Without a home directory, the header shows full paths, so the error is not fatal.
	home, _ := os.UserHomeDir()
	model := ui.New(navigator, ui.Config{Keys: ui.DefaultKeyMap(), Theme: ui.DefaultTheme(), Home: home})
	if _, err := tea.NewProgram(model, tea.WithOutput(os.Stderr)).Run(); err != nil {
		if errors.Is(err, tea.ErrInterrupted) {
			return errCancelled
		}
		return err
	}

	path, ok := model.Result()
	if !ok {
		return errCancelled
	}
	fmt.Println(path)
	return nil
}

func startDir(dir string) (string, error) {
	if dir == "" {
		return os.Getwd()
	}
	return filepath.Abs(dir)
}
