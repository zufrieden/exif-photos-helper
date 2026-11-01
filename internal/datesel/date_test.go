package datesel

import "testing"

func ip(i int) *int { return &i }

func TestEXIF(t *testing.T) {
	tests := []struct {
		name    string
		in      PartialDate
		want    string
		wantErr bool
	}{
		{"decade only", PartialDate{Decade: 1990}, "1990:01:01 00:00:00", false},
		{"year", PartialDate{Decade: 1990, Year: ip(1996)}, "1996:01:01 00:00:00", false},
		{"month", PartialDate{Decade: 1990, Year: ip(1996), Month: ip(7)}, "1996:07:01 00:00:00", false},
		{"day", PartialDate{Decade: 1990, Year: ip(1996), Month: ip(7), Day: ip(14)}, "1996:07:14 00:00:00", false},
		{"hour", PartialDate{Decade: 1990, Year: ip(1996), Month: ip(7), Day: ip(14), Hour: ip(18)}, "1996:07:14 18:00:00", false},
		{"leap day", PartialDate{Decade: 1940, Year: ip(1944), Month: ip(2), Day: ip(29)}, "1944:02:29 00:00:00", false},
		{"bad leap day", PartialDate{Decade: 1940, Year: ip(1943), Month: ip(2), Day: ip(29)}, "", true},
		{"year outside decade", PartialDate{Decade: 1990, Year: ip(2001)}, "", true},
		{"bad decade", PartialDate{Decade: 1995}, "", true},
		{"month without year", PartialDate{Decade: 1990, Month: ip(3)}, "", true},
		{"bad hour", PartialDate{Decade: 1990, Year: ip(1996), Month: ip(7), Day: ip(14), Hour: ip(24)}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.in.EXIF()
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
