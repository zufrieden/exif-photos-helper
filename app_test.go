package main

import (
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"exif-photos-helper/internal/datesel"
	"exif-photos-helper/internal/exif"
)

func TestScanAndApply(t *testing.T) {
	if _, err := exif.Find(); err != nil {
		t.Skip(err)
	}
	t.Setenv("HOME", t.TempDir()) // keep the thumbnail cache out of the real home

	dir := t.TempDir()
	for _, name := range []string{"b.jpg", "A.JPG", "c.jpeg", "notes.txt", ".hidden.jpg", "sub/d.jpg"} {
		p := filepath.Join(dir, name)
		os.MkdirAll(filepath.Dir(p), 0o755)
		f, err := os.Create(p)
		if err != nil {
			t.Fatal(err)
		}
		jpeg.Encode(f, image.NewGray(image.Rect(0, 0, 8, 8)), nil)
		f.Close()
	}

	s := &PhotoService{}
	if s.initErr = s.init(); s.initErr != nil {
		t.Fatal(s.initErr)
	}
	defer s.ServiceShutdown()

	photos, err := s.Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, p := range photos {
		names = append(names, p.Name)
	}
	if len(names) != 3 || names[0] != "A.JPG" || names[1] != "b.jpg" || names[2] != "c.jpeg" {
		t.Fatalf("scanned %v, want [A.JPG b.jpg c.jpeg] (no subfolders, no hidden files)", names)
	}

	year := 1996
	files := []string{photos[0].Path, photos[1].Path}
	updated, err := s.Apply(EditRequest{
		Files:    files,
		Date:     &datesel.PartialDate{Decade: 1990, Year: &year},
		Position: &exif.LatLon{Lat: -12.5, Lon: 130.25},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated) != 2 {
		t.Fatalf("got %d updated photos", len(updated))
	}
	for _, p := range updated {
		if p.Date != "1996:01:01 00:00:00" || p.Lat == nil || *p.Lat != -12.5 || *p.Lon != 130.25 {
			t.Errorf("%s: date=%q lat=%v lon=%v", p.Name, p.Date, p.Lat, p.Lon)
		}
	}

	// Only the date: GPS must stay untouched.
	month := 7
	updated, err = s.Apply(EditRequest{Files: files[:1], Date: &datesel.PartialDate{Decade: 1990, Year: &year, Month: &month}})
	if err != nil {
		t.Fatal(err)
	}
	if updated[0].Date != "1996:07:01 00:00:00" || updated[0].Lat == nil {
		t.Errorf("date-only edit: %+v", updated[0])
	}

	// A missing file reports an error.
	if _, err := s.Apply(EditRequest{Files: []string{filepath.Join(dir, "gone.jpg")}, Position: &exif.LatLon{}}); err == nil {
		t.Error("expected error for missing file")
	}
}
