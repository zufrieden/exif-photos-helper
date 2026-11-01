package exif

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Photo is the metadata the UI needs for one image.
type Photo struct {
	Path        string   `json:"path"`
	Name        string   `json:"name"`
	Date        string   `json:"date"` // "YYYY:MM:DD HH:MM:SS", empty if unknown
	Lat         *float64 `json:"lat"`
	Lon         *float64 `json:"lon"`
	Orientation int      `json:"orientation"`
}

type rawPhoto struct {
	SourceFile       string `json:"SourceFile"`
	FileName         string `json:"FileName"`
	DateTimeOriginal any    `json:"DateTimeOriginal"`
	CreateDate       any    `json:"CreateDate"`
	GPSPosition      any    `json:"GPSPosition"`
	Orientation      any    `json:"Orientation"`
}

// ReadArgs returns exiftool arguments to read metadata of files as JSON.
func ReadArgs(files []string) []string {
	args := []string{"-json", "-n", "-fast2",
		"-FileName", "-DateTimeOriginal", "-CreateDate", "-GPSPosition", "-Orientation"}
	return append(args, files...)
}

// ParsePhotos decodes the output of an exiftool run with ReadArgs.
func ParsePhotos(data []byte) ([]Photo, error) {
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, nil
	}
	var raws []rawPhoto
	if err := json.Unmarshal(data, &raws); err != nil {
		return nil, fmt.Errorf("parse exiftool json: %w", err)
	}
	photos := make([]Photo, 0, len(raws))
	for _, r := range raws {
		p := Photo{Path: r.SourceFile, Name: r.FileName, Orientation: 1}
		p.Date = validDate(r.DateTimeOriginal)
		if p.Date == "" {
			p.Date = validDate(r.CreateDate)
		}
		if lat, lon, ok := parsePosition(r.GPSPosition); ok {
			p.Lat, p.Lon = &lat, &lon
		}
		if o, ok := r.Orientation.(float64); ok && o >= 1 && o <= 8 {
			p.Orientation = int(o)
		}
		photos = append(photos, p)
	}
	return photos, nil
}

// validDate returns v if it is a usable EXIF date string. Cameras without a
// clock often write "0000:00:00 00:00:00", which counts as no date.
func validDate(v any) string {
	s, ok := v.(string)
	if !ok || len(s) < 10 || strings.HasPrefix(s, "0000") {
		return ""
	}
	return s
}

func parsePosition(v any) (lat, lon float64, ok bool) {
	s, isStr := v.(string)
	if !isStr {
		return 0, 0, false
	}
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return 0, 0, false
	}
	lat, err1 := strconv.ParseFloat(parts[0], 64)
	lon, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return lat, lon, true
}

// LatLon is a WGS84 position in signed decimal degrees.
type LatLon struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// WriteArgs builds exiftool arguments that set the date and/or position on
// files, in place and without backup. date is an EXIF date string
// ("YYYY:MM:DD HH:MM:SS"); empty leaves dates untouched. pos nil leaves GPS
// untouched.
func WriteArgs(date string, pos *LatLon, files []string) ([]string, error) {
	if date == "" && pos == nil {
		return nil, fmt.Errorf("nothing to write")
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files")
	}
	args := []string{"-overwrite_original"}
	if date != "" {
		args = append(args,
			"-DateTimeOriginal="+date,
			"-CreateDate="+date,
			"-ModifyDate="+date,
			"-FileModifyDate="+date,
		)
	}
	if pos != nil {
		if math.Abs(pos.Lat) > 90 || math.Abs(pos.Lon) > 180 {
			return nil, fmt.Errorf("invalid position %v,%v", pos.Lat, pos.Lon)
		}
		latRef, lonRef := "N", "E"
		if pos.Lat < 0 {
			latRef = "S"
		}
		if pos.Lon < 0 {
			lonRef = "W"
		}
		args = append(args,
			"-GPSLatitude="+strconv.FormatFloat(math.Abs(pos.Lat), 'f', 7, 64),
			"-GPSLatitudeRef="+latRef,
			"-GPSLongitude="+strconv.FormatFloat(math.Abs(pos.Lon), 'f', 7, 64),
			"-GPSLongitudeRef="+lonRef,
		)
	}
	return append(args, files...), nil
}
