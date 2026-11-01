// Package scan lists the photos directly inside a folder (no subfolders).
package scan

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var extensions = map[string]bool{".jpg": true, ".jpeg": true, ".heic": true, ".heif": true}

// IsPhoto reports whether name has a supported photo extension.
func IsPhoto(name string) bool {
	return extensions[strings.ToLower(filepath.Ext(name))]
}

// List returns the absolute paths of supported photos in dir, sorted by name.
// Hidden files (including macOS "._" resource forks) are skipped.
func List(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || strings.HasPrefix(name, ".") || !IsPhoto(name) {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i]) < strings.ToLower(files[j])
	})
	return files, nil
}
