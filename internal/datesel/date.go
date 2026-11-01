// Package datesel turns a partially-known date (decade → year → month → day
// → hour) into a full EXIF date string. Unknown parts get their lowest value.
package datesel

import (
	"fmt"
	"time"
)

// PartialDate is a date known down to some precision. Decade is required
// (e.g. 1990); nil fields are unknown. A field may only be set if the one
// before it is set.
type PartialDate struct {
	Decade int  `json:"decade"`
	Year   *int `json:"year"`
	Month  *int `json:"month"`
	Day    *int `json:"day"`
	Hour   *int `json:"hour"`
}

// EXIF formats d as "YYYY:MM:DD HH:00:00".
func (d PartialDate) EXIF() (string, error) {
	if d.Decade < 1800 || d.Decade > 2100 || d.Decade%10 != 0 {
		return "", fmt.Errorf("invalid decade %d", d.Decade)
	}
	year, month, day, hour := d.Decade, 1, 1, 0

	if d.Year != nil {
		if *d.Year < d.Decade || *d.Year > d.Decade+9 {
			return "", fmt.Errorf("year %d is not in the %ds", *d.Year, d.Decade)
		}
		year = *d.Year
	}
	if d.Month != nil {
		if d.Year == nil {
			return "", fmt.Errorf("month set without year")
		}
		if *d.Month < 1 || *d.Month > 12 {
			return "", fmt.Errorf("invalid month %d", *d.Month)
		}
		month = *d.Month
	}
	if d.Day != nil {
		if d.Month == nil {
			return "", fmt.Errorf("day set without month")
		}
		if *d.Day < 1 || *d.Day > DaysIn(year, month) {
			return "", fmt.Errorf("invalid day %d for %04d-%02d", *d.Day, year, month)
		}
		day = *d.Day
	}
	if d.Hour != nil {
		if d.Day == nil {
			return "", fmt.Errorf("hour set without day")
		}
		if *d.Hour < 0 || *d.Hour > 23 {
			return "", fmt.Errorf("invalid hour %d", *d.Hour)
		}
		hour = *d.Hour
	}
	return fmt.Sprintf("%04d:%02d:%02d %02d:00:00", year, month, day, hour), nil
}

// DaysIn returns the number of days in the given month.
func DaysIn(year, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
