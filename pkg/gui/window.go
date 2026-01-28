package gui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	fyneDialog "fyne.io/fyne/v2/dialog"
	"gba/pkg/gba"
	nativeDialog "github.com/sqweek/dialog"
)

const (
	GameWidth  = 240
	GameHeight = 160
)

type MainWindow struct {
	app    fyne.App
	window fyne.Window

	emulator *gba.GBA
	canvas   *GameCanvas
	input    *InputHandler

	scale int
}

func NewMainWindow() *MainWindow {
	a := app.New()
	w := a.NewWindow("GBA Emulator")

	mw := &MainWindow{
		app:      a,
		window:   w,
		emulator: gba.New(),
		scale:    2,
	}

	mw.setupUI()

	return mw
}

func (mw *MainWindow) setupUI() {
	// 创建游戏画面渲染区域
	mw.canvas = NewGameCanvas(mw.scale)

	// 创建输入处理器
	mw.input = NewInputHandler(mw.emulator)

	// 创建菜单
	menu := mw.createMenu()
	mw.window.SetMainMenu(menu)

	// 设置内容
	content := container.NewCenter(mw.canvas.Image())
	mw.window.SetContent(content)

	// 设置窗口大小
	mw.updateWindowSize()

	// 设置键盘事件
	mw.setupKeyboardEvents()
}

func (mw *MainWindow) createMenu() *fyne.MainMenu {
	return NewMenuBar(mw).Build()
}

func (mw *MainWindow) setupKeyboardEvents() {
	// 在 canvas 中处理键盘事件
}

func (mw *MainWindow) updateWindowSize() {
	width := float32(GameWidth * mw.scale)
	height := float32(GameHeight * mw.scale)
	mw.window.Resize(fyne.NewSize(width, height))
}

func (mw *MainWindow) SetScale(scale int) {
	if scale < 1 || scale > 4 {
		return
	}
	mw.scale = scale
	mw.canvas.SetScale(scale)
	mw.updateWindowSize()
}

func (mw *MainWindow) LoadROM(filename string) error {
	fmt.Printf("[GUI] Loading ROM from GUI: %s\n", filename)
	err := mw.emulator.LoadROM(filename)
	if err != nil {
		fmt.Printf("[GUI] ERROR: Failed to load ROM: %v\n", err)
		return err
	}
	fmt.Printf("[GUI] ROM loaded successfully, starting game loop\n")
	return nil
}

func (mw *MainWindow) LoadBIOS(filename string) error {
	fmt.Printf("[GUI] Loading BIOS from GUI: %s\n", filename)
	err := mw.emulator.LoadBIOS(filename)
	if err != nil {
		fmt.Printf("[GUI] ERROR: Failed to load BIOS: %v\n", err)
		return err
	}
	fmt.Printf("[GUI] BIOS loaded successfully\n")
	return nil
}

func (mw *MainWindow) Start() {
	// 启动游戏循环
	go mw.gameLoop()

	// 运行 Fyne 应用
	mw.window.ShowAndRun()
}

func (mw *MainWindow) gameLoop() {
	fmt.Printf("[GUI] Game loop started\n")
	frameCount := 0

	for {
		if mw.emulator != nil {
			mw.emulator.RunFrame()

			// 更新画面
			frameBuffer := mw.emulator.GetFrameBuffer()
			mw.canvas.UpdateFrame(frameBuffer)

			frameCount++
			// 每 60 帧输出一次日志
			if frameCount%60 == 0 {
				fmt.Printf("[GUI] Rendered %d frames, FrameBuffer first pixel: 0x%04X\n",
					frameCount, frameBuffer[0])
			}
		}
	}
}

func (mw *MainWindow) ShowOpenFileDialog() {
	// 使用原生 Windows 对话框替代 Fyne 的对话框
	filename, err := nativeDialog.File().
		Title("选择 GBA ROM 文件").
		Filter("GBA ROM 文件", "gba", "GBA").
		Load()

	if err != nil {
		// 用户取消或发生错误
		return
	}

	// 清理文件路径
	filename = filepath.Clean(filename)

	// 加载 ROM
	if err := mw.LoadROM(filename); err != nil {
		fyneDialog.ShowError(err, mw.window)
		return
	}
}

func (mw *MainWindow) Reset() {
	if mw.emulator != nil {
		mw.emulator.Reset()
	}
}

func (mw *MainWindow) ToggleFullscreen() {
	if mw.window.FullScreen() {
		mw.window.SetFullScreen(false)
	} else {
		mw.window.SetFullScreen(true)
	}
}

func (mw *MainWindow) Quit() {
	mw.app.Quit()
}
