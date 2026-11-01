package thumbs

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"exif-photos-helper/internal/exif"
)

func TestGet(t *testing.T) {
	bin, err := exif.Find()
	if err != nil {
		t.Skip(err)
	}
	pool, err := exif.NewPool(bin, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	dir := t.TempDir()
	m, err := New(pool, filepath.Join(dir, "cache"))
	if err != nil {
		t.Fatal(err)
	}

	// A large JPEG without an embedded thumbnail goes through sips.
	src := filepath.Join(dir, "big.jpg")
	img := image.NewRGBA(image.Rect(0, 0, 2000, 1000))
	for x := 0; x < 2000; x++ {
		img.Set(x, x%1000, color.White)
	}
	f, _ := os.Create(src)
	jpeg.Encode(f, img, nil)
	f.Close()
	// Mark it rotated: the thumbnail must come back without the tag.
	if _, _, err := pool.Run("-overwrite_original", "-Orientation#=6", src); err != nil {
		t.Fatal(err)
	}

	for _, size := range []int{Small, Large} {
		data, err := m.Get(src, size)
		if err != nil {
			t.Fatal(err)
		}
		cfg, err := jpeg.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Width != size || cfg.Height != size/2 {
			t.Errorf("size %d: got %dx%d", size, cfg.Width, cfg.Height)
		}
		thumb := filepath.Join(dir, "thumb.jpg")
		os.WriteFile(thumb, data, 0o644)
		out, _, err := pool.Run("-s3", "-Orientation#", thumb)
		if err != nil {
			t.Fatal(err)
		}
		if o := string(bytes.TrimSpace(out)); o != "" && o != "1" {
			t.Errorf("size %d: thumbnail keeps orientation %q", size, o)
		}
	}

	// A HEIC file (converted by sips) also gets a thumbnail.
	heic := filepath.Join(dir, "photo.heic")
	if out, err := exec.Command("/usr/bin/sips", "-s", "format", "heic", src, "--out", heic).CombinedOutput(); err != nil {
		t.Skipf("sips cannot write HEIC: %s", out)
	}
	data, err := m.Get(heic, Small)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, jpegMagic) {
		t.Error("HEIC thumbnail is not a JPEG")
	}

	// Cached: second call must return the same bytes.
	again, _ := m.Get(heic, Small)
	if !bytes.Equal(data, again) {
		t.Error("cache miss on second call")
	}
}
