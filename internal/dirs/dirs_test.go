package dirs

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestListerList(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"b", "A", ".hidden"} {
		mustMkdir(t, filepath.Join(dir, name))
	}
	mustWriteFile(t, filepath.Join(dir, "file.txt"))
	mustSymlink(t, filepath.Join(dir, "b"), filepath.Join(dir, "link"))
	mustSymlink(t, filepath.Join(dir, "file.txt"), filepath.Join(dir, "file-link"))
	mustSymlink(t, filepath.Join(dir, "missing"), filepath.Join(dir, "broken-link"))

	tests := []struct {
		name   string
		lister Lister
		want   []string
	}{
		{name: "hidden excluded", lister: Lister{}, want: []string{"A", "b", "link"}},
		{name: "hidden included", lister: Lister{ShowHidden: true}, want: []string{".hidden", "A", "b", "link"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.lister.List(dir)
			if err != nil {
				t.Fatalf("List() error: %v", err)
			}
			if !slices.Equal(got, test.want) {
				t.Errorf("List() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestListerListEmpty(t *testing.T) {
	got, err := Lister{}.List(t.TempDir())
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("List() = %q, want empty", got)
	}
}

func TestListerListMissingDir(t *testing.T) {
	_, err := Lister{}.List(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("List() error = nil, want error")
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}
