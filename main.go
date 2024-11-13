package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
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
		// 添加窗口居中设置
		StartHidden: true, // 先隐藏窗口
	})

	if err != nil {
		log.Fatal(err)
	}
}
