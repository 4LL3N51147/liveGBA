package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

type MenuBar struct {
	window   *MainWindow
	mainMenu *fyne.MainMenu
}

func NewMenuBar(window *MainWindow) *MenuBar {
	return &MenuBar{
		window: window,
	}
}

func (mb *MenuBar) Build() *fyne.MainMenu {
	// File 菜单
	openItem := fyne.NewMenuItem("Open ROM...", func() {
		mb.window.ShowOpenFileDialog()
	})

	resetItem := fyne.NewMenuItem("Reset", func() {
		mb.window.Reset()
	})

	exitItem := fyne.NewMenuItem("Exit", func() {
		mb.window.Quit()
	})

	fileMenu := fyne.NewMenu("File", openItem, fyne.NewMenuItemSeparator(), resetItem, fyne.NewMenuItemSeparator(), exitItem)

	// View 菜单
	scaleMenu := fyne.NewMenuItem("Scale", nil)
	scale1x := fyne.NewMenuItem("1x", func() {
		mb.window.SetScale(1)
	})
	scale2x := fyne.NewMenuItem("2x", func() {
		mb.window.SetScale(2)
	})
	scale3x := fyne.NewMenuItem("3x", func() {
		mb.window.SetScale(3)
	})
	scale4x := fyne.NewMenuItem("4x", func() {
		mb.window.SetScale(4)
	})
	scaleMenu.ChildMenu = fyne.NewMenu("", scale1x, scale2x, scale3x, scale4x)

	fullscreenItem := fyne.NewMenuItem("Fullscreen", func() {
		mb.window.ToggleFullscreen()
	})

	viewMenu := fyne.NewMenu("View", scaleMenu, fyne.NewMenuItemSeparator(), fullscreenItem)

	// Help 菜单
	aboutItem := fyne.NewMenuItem("About", func() {
		dialog.ShowInformation("About", "GBA Emulator\n\nA Game Boy Advance emulator written in Go with Fyne GUI.", mb.window.window)
	})

	helpMenu := fyne.NewMenu("Help", aboutItem)

	mb.mainMenu = fyne.NewMainMenu(fileMenu, viewMenu, helpMenu)

	return mb.mainMenu
}
