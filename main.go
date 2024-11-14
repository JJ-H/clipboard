package main

import (
	"context"
	"embed"
	"log"
	"sync"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

// 添加全局变量来跟踪退出状态
var (
	isQuitting bool
	quitMutex  sync.Mutex
)

func main() {
	app := NewApp()

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
			quitMutex.Lock()
			defer quitMutex.Unlock()
			
			// 如果是通过 QuitApp 方法退出，允许关闭
			if isQuitting {
				return false
			}
			
			// 否则最小化窗口
			runtime.WindowMinimise(ctx)
			return true
		},
		StartHidden: true,
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title:   "Smart Clipboard",
				Message: "Modern clipboard manager",
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
