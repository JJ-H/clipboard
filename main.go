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
			// 添加日志打印
			log.Println("OnBeforeClose triggered")
			
			// 如果是通过 QuitApp 方法退出，直接退出
			if isQuitting {
				log.Println("Quitting via QuitApp")
				return false
			}
			
			// 如果窗口可见，则隐藏窗口
			if runtime.WindowIsNormal(ctx) {
				log.Println("Window is normal, hiding window")
				runtime.WindowHide(ctx)
				return true
			}
			
			// 如果窗口已经隐藏（从 Dock 菜单退出），直接退出
			log.Println("Window is hidden, quitting app")
			isQuitting = true
			runtime.Quit(ctx)
			return false
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
