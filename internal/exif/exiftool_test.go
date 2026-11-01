package exif

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestWriteArgs(t *testing.T) {
	got, err := WriteArgs("1996:01:01 00:00:00", &LatLon{Lat: -33.8688, Lon: -151.2093}, []string{"a.jpg"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"-overwrite_original",
		"-DateTimeOriginal=1996:01:01 00:00:00", "-CreateDate=1996:01:01 00:00:00",
		"-ModifyDate=1996:01:01 00:00:00", "-FileModifyDate=1996:01:01 00:00:00",
		"-GPSLatitude=33.8688000", "-GPSLatitudeRef=S",
		"-GPSLongitude=151.2093000", "-GPSLongitudeRef=W",
		"a.jpg"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got  %q\nwant %q", got, want)
	}

	if got, _ := WriteArgs("", &LatLon{Lat: 46.5, Lon: 6.6}, []string{"a.jpg"}); got[2] != "-GPSLatitudeRef=N" || got[4] != "-GPSLongitudeRef=E" {
		t.Errorf("north/east refs wrong: %q", got)
	}
	if _, err := WriteArgs("", nil, []string{"a.jpg"}); err == nil {
		t.Error("expected error when nothing to write")
	}
	if _, err := WriteArgs("", &LatLon{Lat: 91}, []string{"a.jpg"}); err == nil {
		t.Error("expected error for invalid latitude")
	}
}

func TestParsePhotos(t *testing.T) {
	data := []byte(`[{"SourceFile":"/x/a.jpg","FileName":"a.jpg","DateTimeOriginal":"0000:00:00 00:00:00","CreateDate":"1996:07:01 00:00:00","GPSPosition":"-33.5 151.25","Orientation":6},
		{"SourceFile":"/x/b.jpg","FileName":"b.jpg"}]`)
	ps, err := ParsePhotos(data)
	if err != nil {
		t.Fatal(err)
	}
	if ps[0].Date != "1996:07:01 00:00:00" || *ps[0].Lat != -33.5 || *ps[0].Lon != 151.25 || ps[0].Orientation != 6 {
		t.Errorf("bad parse: %+v", ps[0])
	}
	if ps[1].Date != "" || ps[1].Lat != nil || ps[1].Orientation != 1 {
		t.Errorf("bad parse: %+v", ps[1])
	}
}

// TestRoundTrip writes to a real JPEG copy through a stay_open exiftool.
func TestRoundTrip(t *testing.T) {
	bin, err := Find()
	if err != nil {
		t.Skip(err)
	}
	src := filepath.Join(t.TempDir(), "photo.jpg")
	writeJPEG(t, src)

	tool, err := Start(bin)
	if err != nil {
		t.Fatal(err)
	}
	defer tool.Close()

	args, _ := WriteArgs("1996:01:01 00:00:00", &LatLon{Lat: 46.5197, Lon: -6.6323}, []string{src})
	out, stderr, err := tool.Run(args...)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "1 image files updated") {
		t.Fatalf("write failed: %s / %s", out, stderr)
	}

	out, _, err = tool.Run(ReadArgs([]string{src})...)
	if err != nil {
		t.Fatal(err)
	}
	ps, err := ParsePhotos(out)
	if err != nil || len(ps) != 1 {
		t.Fatalf("read failed: %v %s", err, out)
	}
	p := ps[0]
	if p.Date != "1996:01:01 00:00:00" {
		t.Errorf("date = %q", p.Date)
	}
	if p.Lat == nil || *p.Lat < 46.5196 || *p.Lat > 46.5198 || *p.Lon > -6.6322 || *p.Lon < -6.6324 {
		t.Errorf("position = %v,%v", p.Lat, p.Lon)
	}
	st, _ := os.Stat(src)
	if st.ModTime().Year() != 1996 {
		t.Errorf("file mtime = %v", st.ModTime())
	}
}

// writeJPEG writes a small solid-colour JPEG fixture.
func writeJPEG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 48))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{200, 120, 40, 255}}, image.Point{}, draw.Src)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, nil); err != nil {
		t.Fatal(err)
	}
}
