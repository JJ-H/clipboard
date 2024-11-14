package main

import (
	"context"
	"embed"
	"log"
	"sync"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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
			// 如果是通过 QuitApp 方法退出，直接退出
			if isQuitting {
				return false
			}

			// log.Println("OnBeforeClose", app.isWindowVisible)

			// 如果是点击窗口关闭按钮，则隐藏窗口
			if app.isWindowVisible {
				runtime.WindowHide(ctx)
				app.isWindowVisible = false // 更新窗口状态
				return true
			}

			// 其他情况（从 Dock 菜单退出），直接退出
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
