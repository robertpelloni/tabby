package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// createAppMenu creates the application menu bar.
func createAppMenu() *menu.Menu {
	appMenu := menu.NewMenu()
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("&New Tab", nil, func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:new-tab")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("&Settings", nil, func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:settings")
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("&Quit", nil, func(_ *menu.CallbackData) {
		wailsRuntime.Quit(app.ctx)
	})

	editMenu := appMenu.AddSubmenu("Edit")
	editMenu.AddText("&Copy", nil, func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:copy")
	})
	editMenu.AddText("&Paste", nil, func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:paste")
	})
	editMenu.AddText("Select &All", nil, func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:select-all")
	})

	viewMenu := appMenu.AddSubmenu("View")
	viewMenu.AddText("Command &Palette", nil, func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:command-palette")
	})
	viewMenu.AddText("&Toggle Full Screen", nil, func(_ *menu.CallbackData) {
		wailsRuntime.WindowToggleMaximise(app.ctx)
	})

	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("&About Tabby Go", nil, func(_ *menu.CallbackData) {
		wailsRuntime.EventsEmit(app.ctx, "menu:about")
	})

	return appMenu
}

var app *App

func main() {
	app = NewApp()

	err := wails.Run(&options.App{
		Title:     "Tabby Go",
		Width:     1280,
		Height:    800,
		MinWidth:  600,
		MinHeight: 400,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 30, G: 30, B: 30, A: 1},
		OnStartup:        app.startup,
		Menu:           createAppMenu(),
		OnBeforeClose:  app.onBeforeClose,
		Bind: []interface{}{
			app,
		},
Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			DisableFramelessWindowDecorations: false,
			// Disable GPU acceleration to prevent WebView2 renderer crashes
			BackdropType: windows.Acrylic,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
