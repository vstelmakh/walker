package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStartDir(t *testing.T) {
	t.Chdir(t.TempDir())
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error: %v", err)
	}

	tests := []struct {
		name string
		dir  string
		want string
	}{
		{name: "empty is the current directory", dir: "", want: wd},
		{name: "relative path", dir: "sub", want: filepath.Join(wd, "sub")},
		{name: "path with a dash", dir: "./-name", want: filepath.Join(wd, "-name")},
		{name: "absolute path", dir: "/tmp", want: "/tmp"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := startDir(test.dir)
			if err != nil {
				t.Fatalf("startDir() error: %v", err)
			}
			if got != test.want {
				t.Errorf("startDir() = %q, want %q", got, test.want)
			}
		})
	}
}
