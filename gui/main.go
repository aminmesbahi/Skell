package main

import (
	"embed"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	app := NewApp()
	appInstance := application.New(application.Options{
		Name:        "Skell",
		Description: "Desktop GUI for the Skell CLI skill manager.",
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		Services: []application.Service{
			application.NewService(app),
		},
	})

	// Set the dock/taskbar icon
	appInstance.SetIcon(appIcon)

	appInstance.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Skell",
		Width:            1320,
		Height:           840,
		MinWidth:         960,
		MinHeight:        600,
		StartState:       application.WindowStateNormal,
		BackgroundColour: application.NewRGBA(10, 13, 26, 255),
		Mac: application.MacWindow{
			TitleBar: application.MacTitleBar{
				HideTitle:       true,
				FullSizeContent: true,
			},
			Appearance: application.NSAppearanceNameDarkAqua,
		},
		Linux: application.LinuxWindow{
			WindowIsTranslucent: false,
		},
	})

	if err := appInstance.Run(); err != nil {
		println("Error:", err.Error())
	}
}
