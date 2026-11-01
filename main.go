package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	// Emitted with the folder path when a folder (or a photo in it) is dropped.
	application.RegisterEvent[string]("folder-dropped")
}

func main() {
	svc := &PhotoService{}

	app := application.New(application.Options{
		Name:        "EXIF Photos Helper",
		Description: "Batch edit photo dates and GPS positions",
		Services: []application.Service{
			application.NewService(svc),
		},
		Assets: application.AssetOptions{
			Handler:    application.AssetFileServerFS(assets),
			Middleware: svc.thumbMiddleware,
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	win := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:          "EXIF Photos Helper",
		Width:          1400,
		Height:         900,
		MinWidth:       900,
		MinHeight:      600,
		EnableFileDrop: true,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 38,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(24, 24, 27),
		URL:              "/",
	})

	win.OnWindowEvent(events.Common.WindowFilesDropped, func(e *application.WindowEvent) {
		files := e.Context().DroppedFiles()
		if len(files) == 0 {
			return
		}
		dir := files[0]
		if st, err := os.Stat(dir); err == nil && !st.IsDir() {
			dir = filepath.Dir(dir)
		}
		app.Event.Emit("folder-dropped", dir)
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
