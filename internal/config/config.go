// Package config resolves the Mapbox access token.
package config

import (
	"bufio"
	"os"
	"strings"
)

// embeddedToken is set at build time from .env:
//
//	-ldflags "-X exif-photos-helper/internal/config.embeddedToken=pk...."
var embeddedToken string

// MapboxToken returns the token from, in order: the MAPBOX_TOKEN environment
// variable, a .env file in the working directory, or the build-time value.
func MapboxToken() string {
	if t := os.Getenv("MAPBOX_TOKEN"); t != "" {
		return t
	}
	if t := readDotEnv(".env", "MAPBOX_TOKEN"); t != "" {
		return t
	}
	return embeddedToken
}

func readDotEnv(path, key string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(k) == key {
			return strings.Trim(strings.TrimSpace(v), `"'`)
		}
	}
	return ""
}
