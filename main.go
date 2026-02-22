package main

import (
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	a "dbmx/app"
	"dbmx/config/database"
	"dbmx/config/env"
)

//go:embed all:frontend/build
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	// Create an instance of the app structure
	env := env.GetEnv()
	db, err := database.NewSqlite3DB(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer db.CloseConn()

	pm := a.NewPoolManager()

	conn := a.NewConnections(db.DB, pm)
	tabs := a.NewTabs(db.DB, pm)
	auth := a.InitAuth(db.DB, env.SupabaseConfig)
	app := NewApp(conn)

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "dbmx",
		Width:  1024,
		Height: 768,
		// MinWidth:  1024,
		// MinHeight: 768,
		// MaxWidth:          1280,
		// MaxHeight:         800,
		DisableResize: false,
		Fullscreen:    false,
		// Frameless:         runtime.GOOS != "darwin",
		StartHidden:       false,
		HideWindowOnClose: false,
		// BackgroundColor:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Handler:    nil,
			Middleware: nil,
		},
		Menu:             nil,
		Logger:           nil,
		LogLevel:         logger.WARNING,
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		WindowStartState: options.Normal,
		Bind: []interface{}{
			app,
			conn,
			tabs,
			auth,
		},
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHidden(),
			Appearance:           mac.NSAppearanceNameAccessibilityHighContrastVibrantLight,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "dbmx",
				Message: "",
				Icon:    icon,
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

}
