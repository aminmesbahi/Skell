package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "Skell",
		Width:            1320,
		Height:           840,
		MinWidth:         960,
		MinHeight:        600,
		DisableResize:    false,
		Fullscreen:       false,
		WindowStartState: options.Normal,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// Alpha must be 255 (fully opaque). With A: 1 the native Win32 window
		// is effectively transparent under WebView2, so each route transition
		// flashes the underlying white window through the webview repaint —
		// visible as flicker when navigating between pages on Windows.
		BackgroundColour: &options.RGBA{R: 10, G: 13, B: 26, A: 255},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisablePinchZoom:     false,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
			// Force the dark appearance so macOS does not paint its
			// light 1px highlight line at the top of the window.
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "Skell",
				Message: "Desktop GUI for the Skell CLI skill manager.",
			},
		},
		Linux: &linux.Options{
			WindowIsTranslucent: false,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
