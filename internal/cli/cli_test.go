package cli

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Options
		wantErr string
	}{
		{name: "no arguments", args: nil, want: Options{}},
		{name: "directory", args: []string{"projects"}, want: Options{Dir: "projects"}},
		{name: "help", args: []string{"--help"}, want: Options{Help: true}},
		{name: "help short", args: []string{"-h"}, want: Options{Help: true}},
		{name: "version", args: []string{"--version"}, want: Options{Version: true}},
		{name: "version short", args: []string{"-v"}, want: Options{Version: true}},
		{name: "short options combined", args: []string{"-vh"}, want: Options{Help: true, Version: true}},
		{name: "option before directory", args: []string{"-v", "projects"}, want: Options{Version: true, Dir: "projects"}},
		{name: "option after directory", args: []string{"projects", "-v"}, want: Options{Version: true, Dir: "projects"}},
		{name: "directory after end of options", args: []string{"--", "-name"}, want: Options{Dir: "-name"}},
		{name: "directory with path", args: []string{"./-name"}, want: Options{Dir: "./-name"}},
		{name: "unknown option", args: []string{"--unknown"}, wantErr: "unknown flag: --unknown"},
		{name: "unknown short option", args: []string{"-x"}, wantErr: "unknown shorthand flag: 'x' in -x"},
		{name: "long option with single dash", args: []string{"-version"}, wantErr: "unknown shorthand flag"},
		{name: "too many arguments", args: []string{"a", "b"}, wantErr: "usage: walker [options] [dir]"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Parse(test.args)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("Parse() error = %v, want it to contain %q", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}
			if got != test.want {
				t.Errorf("Parse() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestParseErrorHint(t *testing.T) {
	_, err := Parse([]string{"--hepl"})
	if err == nil || !strings.HasSuffix(err.Error(), "see: walker --help") {
		t.Fatalf("Parse() error = %v, want it to end with the help hint", err)
	}
}

func TestOptionsInfo(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want string
	}{
		{name: "nothing", opts: Options{}, want: ""},
		{name: "directory", opts: Options{Dir: "projects"}, want: ""},
		{name: "help", opts: Options{Help: true}, want: helpText()},
		{name: "version", opts: Options{Version: true}, want: "walker 1.2.3\n"},
		{name: "help wins over version", opts: Options{Help: true, Version: true}, want: helpText()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.opts.Info("1.2.3"); got != test.want {
				t.Errorf("Info() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestHelpText(t *testing.T) {
	got := helpText()
	for _, want := range []string{"Usage: walker [options] [dir]", "Options:", "-h, --help", "-v, --version"} {
		if !strings.Contains(got, want) {
			t.Errorf("helpText() = %q, want it to contain %q", got, want)
		}
	}
}
