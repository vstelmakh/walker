package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/vstelmakh/walker/internal/dirs"
	"github.com/vstelmakh/walker/internal/nav"
	"github.com/vstelmakh/walker/internal/search"
	"github.com/vstelmakh/walker/internal/ui"
)

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
	start, err := startDir(args)
	if err != nil {
		return err
	}

	navigator, err := nav.New(dirs.Lister{ShowHidden: false}, search.Matcher{}, start)
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

func startDir(args []string) (string, error) {
	switch len(args) {
	case 0:
		return os.Getwd()
	case 1:
		return filepath.Abs(args[0])
	default:
		return "", errors.New("usage: walker [dir]")
	}
}
