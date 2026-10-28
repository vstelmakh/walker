// Package cli reads the command line options and builds the help and version texts.
package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/pflag"

	"github.com/vstelmakh/walker/internal/config"
)

type Options struct {
	Help    bool
	Version bool
	// Dir is the directory to start in. It is empty for the current directory.
	Dir string
}

// Parse reads the options and the optional directory. Use -- before a directory that starts with a dash.
// The options given on the command line override the values in cfg. The other values stay as they are.
func Parse(args []string, cfg config.Config) (Options, config.Config, error) {
	var opts Options
	set := newFlagSet(&opts, &cfg)
	if err := set.Parse(args); err != nil {
		return Options{}, config.Config{}, fmt.Errorf("%w, see: walker --help", err)
	}

	switch set.NArg() {
	case 0:
	case 1:
		opts.Dir = set.Arg(0)
	default:
		return Options{}, config.Config{}, errors.New("usage: walker [options] [dir]")
	}
	return opts, cfg, nil
}

// Info returns the text to print instead of starting the UI, or an empty string when there is none.
// The help wins over the version.
func (o Options) Info(version string) string {
	switch {
	case o.Help:
		return helpText()
	case o.Version:
		return fmt.Sprintln("walker", version)
	default:
		return ""
	}
}

// newFlagSet registers every option, one line each. The help text is generated from it.
func newFlagSet(opts *Options, cfg *config.Config) *pflag.FlagSet {
	set := pflag.NewFlagSet("walker", pflag.ContinueOnError)
	set.SetOutput(io.Discard) // the caller reports the error
	set.BoolVarP(&opts.Help, "help", "h", false, "Show this help")
	set.BoolVarP(&opts.Version, "version", "v", false, "Show the version")
	set.BoolVar(&cfg.ShowHidden, "show-hidden", cfg.ShowHidden, "Show hidden directories")
	return set
}

func helpText() string {
	cfg := config.Default()
	return `Walker - terminal directory navigator.

Usage: walker [options] [dir]

Starts in [dir], or in the current directory. Prints the selected directory,
so it can be used like this: cd "$(walker)"

See https://github.com/vstelmakh/walker for more details.

Options:
` + newFlagSet(&Options{}, &cfg).FlagUsages()
}
