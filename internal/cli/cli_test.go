package cli

import (
	"strings"
	"testing"

	"github.com/vstelmakh/walker/internal/config"
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
		{name: "setting with directory", args: []string{"--show-hidden", "projects"}, want: Options{Dir: "projects"}},
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
			got, _, err := Parse(test.args, config.Default())
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

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
		args []string
		want config.Config
	}{
		{
			name: "no options keep defaults",
			cfg:  config.Config{},
			args: nil,
			want: config.Config{},
		},
		{
			name: "no options keep given config",
			cfg:  config.Config{ShowHidden: true},
			args: nil,
			want: config.Config{ShowHidden: true},
		},
		{
			name: "show hidden",
			cfg:  config.Config{},
			args: []string{"--show-hidden"},
			want: config.Config{ShowHidden: true},
		},
		{
			name: "show hidden false overrides config",
			cfg:  config.Config{ShowHidden: true},
			args: []string{"--show-hidden=false"},
			want: config.Config{ShowHidden: false},
		},
		{
			name: "last option wins",
			cfg:  config.Config{},
			args: []string{"--show-hidden", "--show-hidden=false"},
			want: config.Config{ShowHidden: false},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, got, err := Parse(test.args, test.cfg)
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}
			if got != test.want {
				t.Errorf("Parse() config = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestParseErrorHint(t *testing.T) {
	_, _, err := Parse([]string{"--hepl"}, config.Default())
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
	wants := []string{"Usage: walker [options] [dir]", "Options:", "-h, --help", "-v, --version", "    --show-hidden"}
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Errorf("helpText() = %q, want it to contain %q", got, want)
		}
	}
}
