package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"exif-photos-helper/internal/config"
	"exif-photos-helper/internal/datesel"
	"exif-photos-helper/internal/exif"
	"exif-photos-helper/internal/scan"
	"exif-photos-helper/internal/thumbs"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// PhotoService is the API exposed to the frontend.
type PhotoService struct {
	tool    *exif.Tool // metadata reads and writes
	pool    *exif.Pool // thumbnail extraction
	thumbs  *thumbs.Maker
	initErr error
}

// ServiceStartup launches the exiftool processes.
func (s *PhotoService) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	s.initErr = s.init()
	return nil // reported to the UI by Status instead of aborting startup
}

func (s *PhotoService) init() error {
	bin, err := exif.Find()
	if err != nil {
		return err
	}
	if s.tool, err = exif.Start(bin); err != nil {
		return err
	}
	if s.pool, err = exif.NewPool(bin, 4); err != nil {
		return err
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	s.thumbs, err = thumbs.New(s.pool, filepath.Join(cacheDir, "exif-photos-helper", "thumbs"))
	return err
}

// ServiceShutdown stops the exiftool processes.
func (s *PhotoService) ServiceShutdown() error {
	if s.tool != nil {
		_ = s.tool.Close()
	}
	if s.pool != nil {
		s.pool.Close()
	}
	return nil
}

// Status returns an error message if the app can't work (e.g. no exiftool).
func (s *PhotoService) Status() string {
	if s.initErr != nil {
		return s.initErr.Error()
	}
	return ""
}

// InitialFolder returns the folder passed on the command line, if any.
func (s *PhotoService) InitialFolder() string {
	if len(os.Args) > 1 {
		if st, err := os.Stat(os.Args[1]); err == nil && st.IsDir() {
			abs, _ := filepath.Abs(os.Args[1])
			return abs
		}
	}
	return ""
}

// MapboxToken returns the Mapbox access token.
func (s *PhotoService) MapboxToken() string {
	return config.MapboxToken()
}

// ChooseFolder opens a native folder picker. Returns "" if cancelled.
func (s *PhotoService) ChooseFolder() (string, error) {
	return application.Get().Dialog.OpenFile().
		SetTitle("Choose a photo folder").
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()
}

// Scan reads the photos directly inside dir.
func (s *PhotoService) Scan(dir string) ([]exif.Photo, error) {
	if s.initErr != nil {
		return nil, s.initErr
	}
	files, err := scan.List(dir)
	if err != nil {
		return nil, err
	}
	return s.read(files)
}

func (s *PhotoService) read(files []string) ([]exif.Photo, error) {
	if len(files) == 0 {
		return []exif.Photo{}, nil
	}
	out, _, err := s.tool.Run(exif.ReadArgs(files)...)
	if err != nil {
		return nil, err
	}
	return exif.ParsePhotos(out)
}

// DatePreview returns the EXIF date string a partial date will be written as.
func (s *PhotoService) DatePreview(d datesel.PartialDate) (string, error) {
	return d.EXIF()
}

// EditRequest describes a batch edit. Date and Position are each optional.
type EditRequest struct {
	Files    []string             `json:"files"`
	Date     *datesel.PartialDate `json:"date"`
	Position *exif.LatLon         `json:"position"`
}

var updatedRe = regexp.MustCompile(`(\d+) image files? updated`)

// Apply writes the date and/or position to all files, in place, then returns
// their refreshed metadata.
func (s *PhotoService) Apply(req EditRequest) ([]exif.Photo, error) {
	if s.initErr != nil {
		return nil, s.initErr
	}
	date := ""
	if req.Date != nil {
		var err error
		if date, err = req.Date.EXIF(); err != nil {
			return nil, err
		}
	}
	args, err := exif.WriteArgs(date, req.Position, req.Files)
	if err != nil {
		return nil, err
	}
	out, stderr, err := s.tool.Run(args...)
	if err != nil {
		return nil, err
	}

	updated := 0
	if m := updatedRe.FindStringSubmatch(string(out)); m != nil {
		fmt.Sscan(m[1], &updated)
	}
	photos, readErr := s.read(req.Files)
	if updated != len(req.Files) {
		msg := strings.TrimSpace(stderr)
		if msg == "" {
			msg = strings.TrimSpace(string(out))
		}
		return photos, fmt.Errorf("%d of %d files updated: %s", updated, len(req.Files), msg)
	}
	return photos, readErr
}

// thumbMiddleware serves /thumb?path=… with a JPEG preview.
func (s *PhotoService) thumbMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/thumb" {
			next.ServeHTTP(w, r)
			return
		}
		path := r.URL.Query().Get("path")
		if s.thumbs == nil || !filepath.IsAbs(path) || !scan.IsPhoto(path) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		size := thumbs.Small
		if r.URL.Query().Get("size") == "large" {
			size = thumbs.Large
		}
		data, err := s.thumbs.Get(path, size)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, os.ErrNotExist) {
				status = http.StatusNotFound
			}
			http.Error(w, err.Error(), status)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(data)
	})
}
