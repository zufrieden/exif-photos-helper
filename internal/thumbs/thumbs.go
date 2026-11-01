// Package thumbs produces small JPEG previews of photos, cached on disk.
package thumbs

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"exif-photos-helper/internal/exif"
)

// Longest edge, in pixels, of grid thumbnails and of large previews.
const (
	Small = 360
	Large = 1600
)

var jpegMagic = []byte{0xFF, 0xD8, 0xFF}

// Maker returns thumbnails, generating and caching them as needed.
type Maker struct {
	pool     *exif.Pool
	cacheDir string
	sips     chan struct{} // limits concurrent sips processes
}

// New creates a Maker caching into cacheDir.
func New(pool *exif.Pool, cacheDir string) (*Maker, error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, err
	}
	return &Maker{pool: pool, cacheDir: cacheDir, sips: make(chan struct{}, 4)}, nil
}

// Get returns JPEG bytes for path, with its longest edge around size pixels
// (Small or Large). The returned image is NOT rotated; the caller applies the
// EXIF orientation.
func (m *Maker) Get(path string, size int) ([]byte, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	cached := filepath.Join(m.cacheDir, cacheKey(path, size, st.Size(), st.ModTime().UnixNano())+".jpg")
	if data, err := os.ReadFile(cached); err == nil {
		return data, nil
	}

	var data []byte
	if size <= Small {
		data, _ = m.embedded(path)
	}
	if len(data) == 0 {
		if data, err = m.resize(path, size); err != nil {
			return nil, err
		}
	}
	_ = os.WriteFile(cached, data, 0o644)
	return data, nil
}

// embedded extracts the thumbnail stored in the EXIF data (fast, ~160px).
func (m *Maker) embedded(path string) ([]byte, error) {
	out, _, err := m.pool.Run("-b", "-ThumbnailImage", path)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(out, jpegMagic) {
		return nil, nil
	}
	return out, nil
}

// resize renders a thumbnail with macOS sips (handles JPEG and HEIC). The
// orientation tag is removed so all thumbnails are unrotated.
func (m *Maker) resize(path string, size int) ([]byte, error) {
	m.sips <- struct{}{}
	defer func() { <-m.sips }()

	tmp, err := os.CreateTemp(m.cacheDir, "sips-*.jpg")
	if err != nil {
		return nil, err
	}
	tmpName := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpName)

	cmd := exec.Command("/usr/bin/sips", "-Z", strconv.Itoa(size),
		"-s", "format", "jpeg", "-s", "formatOptions", "70", path, "--out", tmpName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("sips: %v: %s", err, out)
	}
	if _, _, err := m.pool.Run("-overwrite_original", "-Orientation=", tmpName); err != nil {
		return nil, err
	}
	return os.ReadFile(tmpName)
}

func cacheKey(path string, edge int, size, mtime int64) string {
	h := sha1.Sum([]byte(fmt.Sprintf("%s|%d|%d|%d", path, edge, size, mtime)))
	return hex.EncodeToString(h[:])
}
