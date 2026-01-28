package gui

import (
	"gba/pkg/gba"
	"gba/pkg/input"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

type InputHandler struct {
	emulator *gba.GBA
	keys     map[fyne.KeyName]int
}

func NewInputHandler(emulator *gba.GBA) *InputHandler {
	ih := &InputHandler{
		emulator: emulator,
		keys:     make(map[fyne.KeyName]int),
	}

	ih.setupKeyMap()

	return ih
}

func (ih *InputHandler) setupKeyMap() {
	// 键盘映射到 GBA 按键
	ih.keys[fyne.KeyZ] = input.KeyA
	ih.keys[fyne.KeyX] = input.KeyB
	ih.keys[fyne.KeyReturn] = input.KeyStart
	ih.keys["LeftShift"] = input.KeySelect
	ih.keys["RightShift"] = input.KeySelect
	ih.keys[fyne.KeyUp] = input.KeyUp
	ih.keys[fyne.KeyDown] = input.KeyDown
	ih.keys[fyne.KeyLeft] = input.KeyLeft
	ih.keys[fyne.KeyRight] = input.KeyRight
	ih.keys[fyne.KeyA] = input.KeyL
	ih.keys[fyne.KeyS] = input.KeyR
}

func (ih *InputHandler) HandleKeyDown(key *fyne.KeyEvent) {
	if gbaKey, ok := ih.keys[key.Name]; ok {
		ih.emulator.SetKey(gbaKey, true)
	}
}

func (ih *InputHandler) HandleKeyUp(key *fyne.KeyEvent) {
	if gbaKey, ok := ih.keys[key.Name]; ok {
		ih.emulator.SetKey(gbaKey, false)
	}
}

func (ih *InputHandler) SetupKeyboard(canvas fyne.Canvas) {
	// 设置自定义键盘事件处理
	if desktopCanvas, ok := canvas.(desktop.Canvas); ok {
		desktopCanvas.SetOnKeyDown(func(key *fyne.KeyEvent) {
			ih.HandleKeyDown(key)
		})
		desktopCanvas.SetOnKeyUp(func(key *fyne.KeyEvent) {
			ih.HandleKeyUp(key)
		})
	}
}

func (ih *InputHandler) IsKeyPressed(keyName fyne.KeyName) bool {
	// 这个方法可以通过 Fyne 的键盘状态来检查
	// 但目前 Fyne 不提供全局键盘状态查询
	return false
}
