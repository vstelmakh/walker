// Package dirs lists directories on the OS file system.
package dirs

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Lister struct {
	ShowHidden bool
}

// List sorts the names case-insensitively and includes symbolic links to directories.
func (l Lister) List(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !l.ShowHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if !isDir(dir, entry) {
			continue
		}
		names = append(names, entry.Name())
	}

	slices.SortFunc(names, compareNames)
	return names, nil
}

func isDir(dir string, entry fs.DirEntry) bool {
	if entry.IsDir() {
		return true
	}
	if entry.Type()&fs.ModeSymlink == 0 {
		return false
	}
	info, err := os.Stat(filepath.Join(dir, entry.Name()))
	return err == nil && info.IsDir()
}

func compareNames(a, b string) int {
	if result := strings.Compare(strings.ToLower(a), strings.ToLower(b)); result != 0 {
		return result
	}
	return strings.Compare(a, b)
}
