package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	hook "github.com/robotn/gohook"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	xclip "golang.design/x/clipboard"
)

// Tag
type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// ClipboardItem
type ClipboardItem struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	TagID     string    `json:"tagId"`
	Timestamp time.Time `json:"timestamp"`
	size      int       // 添加大小字段，但不序列化
}

// Config
type Config struct {
	MaxHistory int   `json:"maxHistory"`
	Tags       []Tag `json:"tags"`
	AutoHide   bool  `json:"autoHide"`
}

// App struct
type App struct {
	ctx             context.Context
	history         []ClipboardItem
	config          Config
	stop            chan bool
	mutex           sync.Mutex
	skipNextWatch   bool
	isWindowVisible bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		history: make([]ClipboardItem, 0),
		config: Config{
			MaxHistory: 50,
		},
		stop: make(chan bool),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.loadConfig()
	a.loadHistory()
	a.isWindowVisible = true

	// 初始化剪贴板
	err := xclip.Init()
	if err != nil {
		fmt.Printf("initial clipboard failed: %v\n", err)
	}

	// 窗口居中显示
	runtime.WindowCenter(ctx)
	runtime.WindowShow(ctx)

	// 设置窗口关闭事件处理
	runtime.EventsOn(ctx, "window-close-requested", func(optionalData ...interface{}) {
		// 阻止窗口关闭，改为最小化
		runtime.WindowMinimise(ctx)
	})

	// 启动剪贴板监听
	go a.watchClipboard()
	// 启动快捷键监听
	go a.watchHotkey()
}

// 添加快捷键监听
func (a *App) watchHotkey() {
	hook.Register(hook.KeyDown, []string{"shift", "ctrl", "v"}, func(e hook.Event) {
		runtime.EventsEmit(a.ctx, "toggleWindow")
	})

	hook.Register(hook.KeyDown, []string{"alt", "c"}, func(e hook.Event) {
		runtime.EventsEmit(a.ctx, "toggleWindow")
	})

	s := hook.Start()
	<-hook.Process(s)
}

// watchClipboard 监听剪贴板变化
func (a *App) watchClipboard() {
	var lastContent string

	// 监听文本变化
	textChan := xclip.Watch(a.ctx, xclip.FmtText)
	// 监听图片变化
	imageChan := xclip.Watch(a.ctx, xclip.FmtImage)

	for {
		select {
		case <-a.stop:
			return
		case data := <-textChan:
			if data != nil && !a.skipNextWatch {
				content := string(data)
				if content != lastContent && content != "" {
					lastContent = content
					go a.saveClipboardItem(content, "text")
				}
			}
			a.skipNextWatch = false
		case data := <-imageChan:
			if data != nil && !a.skipNextWatch {
				// 检查图片大小
				if len(data) > 10*1024*1024 { // 10MB 限制
					runtime.EventsEmit(a.ctx, "clipboardError", "图片大小超过限制(10MB)")
					continue
				}

				imgBase64 := base64.StdEncoding.EncodeToString(data)
				imgContent := "data:image/png;base64," + imgBase64
				if imgContent != lastContent {
					lastContent = imgContent
					go a.saveClipboardItem(imgContent, "image")
				}
			}
			a.skipNextWatch = false
		}
	}
}

// saveClipboardItem
func (a *App) saveClipboardItem(content string, itemType string) {
	// 如果是图片，进行压缩处理
	if itemType == "image" {
		// 限制base64字符串的最大长度（约1MB）
		if len(content) > 1024*1024 {
			// 简单的裁剪，实际项目中可以使用图片压缩库
			content = content[:1024*1024]
		}
	}

	item := ClipboardItem{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Content:   content,
		Type:      itemType,
		Timestamp: time.Now(),
	}

	a.mutex.Lock()
	a.history = append([]ClipboardItem{item}, a.history...)
	if len(a.history) > a.config.MaxHistory {
		a.history = a.history[:a.config.MaxHistory]
	}
	a.mutex.Unlock()

	// 异步保存
	go a.saveHistory()
	runtime.EventsEmit(a.ctx, "historyUpdated")
}

// ToggleWindow 切换窗口显示状态
func (a *App) ToggleWindow() {

	if a.isWindowVisible {
		// 如果窗口当前可见，则隐藏
		runtime.WindowHide(a.ctx)
		a.isWindowVisible = false
	} else {
		// 如果窗口当前不可见，则显示
		runtime.WindowShow(a.ctx)
		runtime.WindowSetAlwaysOnTop(a.ctx, true)
		runtime.WindowSetAlwaysOnTop(a.ctx, false)
		runtime.WindowCenter(a.ctx)
		a.isWindowVisible = true
	}
}

// GetHistory 获取剪贴板历史记录
func (a *App) GetHistory() []ClipboardItem {
	return a.history
}

// SaveToClipboard 保存内容到剪贴板
func (a *App) SaveToClipboard(content string) error {
	a.skipNextWatch = true

	// 如果是图片内容
	if len(content) > 23 && content[:22] == "data:image/png;base64," {
		imgData, err := base64.StdEncoding.DecodeString(content[22:])
		if err != nil {
			return fmt.Errorf("failed to decode image: %w", err)
		}
		xclip.Write(xclip.FmtImage, imgData)
		return nil
	}

	// 如果是文本内容
	xclip.Write(xclip.FmtText, []byte(content))
	return nil
}

// UpdateConfig 更新配置
func (a *App) UpdateConfig(maxHistory int, autoHide bool) error {
	// 添加最大值验证
	if maxHistory > 50 {
		return fmt.Errorf("最大历史记录数不能超过 50")
	}
	a.config.MaxHistory = maxHistory
	a.config.AutoHide = autoHide
	return a.saveConfig()
}

// GetConfig 获取当前配置
func (a *App) GetConfig() Config {
	return a.config
}

// 获取配置文件路径
func (a *App) getConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "clipboard", "config.json")
}

// 获取历史记录文件路径
func (a *App) getHistoryPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "clipboard", "history.json")
}

// 保存配置到文件
func (a *App) saveConfig() error {
	configPath := a.getConfigPath()
	os.MkdirAll(filepath.Dir(configPath), 0755)

	data, err := json.Marshal(a.config)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// 加载配置
func (a *App) loadConfig() {
	data, err := os.ReadFile(a.getConfigPath())
	if err != nil {
		return // 使用默认配置
	}
	json.Unmarshal(data, &a.config)
}

// 保存历史记录到文件
func (a *App) saveHistory() error {
	historyPath := a.getHistoryPath()
	os.MkdirAll(filepath.Dir(historyPath), 0755)

	data, err := json.Marshal(a.history)
	if err != nil {
		return err
	}
	return os.WriteFile(historyPath, data, 0644)
}

// 加载历史记录
func (a *App) loadHistory() {
	data, err := os.ReadFile(a.getHistoryPath())
	if err != nil {
		return
	}
	json.Unmarshal(data, &a.history)
}

// shutdown 方法用于清理
func (a *App) shutdown(ctx context.Context) {
	close(a.stop)
	hook.End()
}

// DeleteHistoryItem 删除指定的历史记录
func (a *App) DeleteHistoryItem(id string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// 查找并删除指定 ID 的记录
	for i, item := range a.history {
		if item.ID == id {
			// 从切片中删除该元素
			a.history = append(a.history[:i], a.history[i+1:]...)
			// 保存更新后的历史记录
			return a.saveHistory()
		}
	}

	return fmt.Errorf("item not found")
}

// AddTag 添加标签
func (a *App) AddTag(name, color string) error {
	tag := Tag{
		ID:    fmt.Sprintf("tag_%d", time.Now().UnixNano()),
		Name:  name,
		Color: color,
	}
	a.config.Tags = append(a.config.Tags, tag)
	return a.saveConfig()
}

// UpdateItemTag 更新项目标签
func (a *App) UpdateItemTag(itemID, tagID string) error {
	for i, item := range a.history {
		if item.ID == itemID {
			a.history[i].TagID = tagID
			return a.saveHistory()
		}
	}
	return fmt.Errorf("item not found")
}

// DeleteTag 删除标签及其关联内容
func (a *App) DeleteTag(tagID string) error {
	// 找到并删除标签
	var found bool
	for i, tag := range a.config.Tags {
		if tag.ID == tagID {
			a.config.Tags = append(a.config.Tags[:i], a.config.Tags[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("tag not found")
	}

	// 删除带有该标签的所有史记录
	var newHistory []ClipboardItem
	for _, item := range a.history {
		if item.TagID != tagID {
			newHistory = append(newHistory, item)
		}
	}
	a.history = newHistory

	// 保存更改
	if err := a.saveHistory(); err != nil {
		return err
	}
	return a.saveConfig()
}

// UpdateTag 更新标签
func (a *App) UpdateTag(id, name, color string) error {
	for i, tag := range a.config.Tags {
		if tag.ID == id {
			a.config.Tags[i].Name = name
			a.config.Tags[i].Color = color
			return a.saveConfig()
		}
	}
	return fmt.Errorf("tag not found")
}

// UpdateTagsOrder 更新标签顺序
func (a *App) UpdateTagsOrder(tagIDs []string) error {
	if len(tagIDs) != len(a.config.Tags) {
		return fmt.Errorf("invalid tags count")
	}

	// 创建新的标签数组
	newTags := make([]Tag, len(tagIDs))
	tagMap := make(map[string]Tag)

	// 创建标签ID到标签的映射
	for _, tag := range a.config.Tags {
		tagMap[tag.ID] = tag
	}

	// 按新顺序重组标签
	for i, id := range tagIDs {
		tag, exists := tagMap[id]
		if !exists {
			return fmt.Errorf("tag not found: %s", id)
		}
		newTags[i] = tag
	}

	a.config.Tags = newTags
	return a.saveConfig()
}

// MoveItemToFront 将指定项目移动到最前面
func (a *App) MoveItemToFront(id string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// 查找指定项目
	var targetItem ClipboardItem
	var targetIndex int = -1

	for i, item := range a.history {
		if item.ID == id {
			targetItem = item
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		return fmt.Errorf("item not found")
	}

	// 如果已经在最前面，不需要移动
	if targetIndex == 0 {
		return nil
	}

	// 移除原位置的项目
	a.history = append(a.history[:targetIndex], a.history[targetIndex+1:]...)
	// 添加到最前面
	a.history = append([]ClipboardItem{targetItem}, a.history...)

	return a.saveHistory()
}

func (a *App) cleanup() {
	// 定期清理过期的历史记录
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			select {
			case <-a.stop:
				ticker.Stop()
				return
			case <-ticker.C:
				a.mutex.Lock()
				// 清理超过一定时间的记录
				now := time.Now()
				var newHistory []ClipboardItem
				for _, item := range a.history {
					if now.Sub(item.Timestamp) < 24*time.Hour {
						newHistory = append(newHistory, item)
					}
				}
				a.history = newHistory
				a.mutex.Unlock()
				a.saveHistory()
			}
		}
	}()
}

// MinimizeWindow 最小化窗口
func (a *App) MinimizeWindow() {
	runtime.WindowMinimise(a.ctx)
}

// QuitApp 直接退出应用
func (a *App) QuitApp() {
	quitMutex.Lock()
	isQuitting = true
	quitMutex.Unlock()
	runtime.Quit(a.ctx)
}

// HideWindow 隐藏窗口
func (a *App) HideWindow() {
	runtime.WindowHide(a.ctx)
	a.isWindowVisible = false
}

// ShowWindow 显示窗口
func (a *App) ShowWindow() {
	runtime.WindowShow(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
	runtime.WindowCenter(a.ctx)
	a.isWindowVisible = true
}
