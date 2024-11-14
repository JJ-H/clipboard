package main

import (
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 创建一个新的应用实例
	app := NewApp()

	// 创建应用配置
	err := wails.Run(&options.App{
		Title:      "📋 Smart Clipboard",
		Width:      800,
		Height:     320,
		Assets:     assets,
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WindowIsTranslucent:  true,
			WebviewIsTransparent: true,
		},
		OnBeforeClose: func(ctx context.Context) bool {
			// 返回 true 阻止窗口关闭，改为最小化
			runtime.WindowMinimise(ctx)
			return true
		},
		StartHidden: true,
	})

	if err != nil {
		log.Fatal(err)
	}
}
