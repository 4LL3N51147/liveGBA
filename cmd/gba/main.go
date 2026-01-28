package main

import (
	"flag"
	"fmt"
	"gba/pkg/gui"
	"os"
)

func main() {
	// 定义命令行参数
	biosFile := flag.String("bios", "", "Path to GBA BIOS file (optional)")
	flag.Parse()

	// 创建主窗口
	window := gui.NewMainWindow()

	// 如果提供了 BIOS 文件，先加载
	if *biosFile != "" {
		fmt.Printf("Loading BIOS: %s\n", *biosFile)
		if err := window.LoadBIOS(*biosFile); err != nil {
			fmt.Printf("Warning: Failed to load BIOS: %v\n", err)
			fmt.Println("Continuing without BIOS...")
		} else {
			fmt.Printf("BIOS loaded successfully\n")
		}
	}

	// 处理剩余参数（ROM 文件）
	args := flag.Args()
	if len(args) >= 1 {
		romFile := args[0]
		// 跳过帮助参数
		if romFile == "--help" || romFile == "-h" {
			printHelp()
			os.Exit(0)
		}

		if err := window.LoadROM(romFile); err != nil {
			fmt.Printf("Failed to load ROM: %v\n", err)
			// 继续运行，用户可以从菜单打开文件
		} else {
			fmt.Printf("Loaded ROM: %s\n", romFile)
		}
	}

	// 启动 GUI
	window.Start()
}

func printHelp() {
	fmt.Println("GBA Emulator with Fyne GUI")
	fmt.Println("Usage: gba [options] <rom_file.gba>")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -bios string    Path to GBA BIOS file (optional)")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  gba game.gba")
	fmt.Println("  gba -bios gba_bios.bin game.gba")
	fmt.Println("")
	fmt.Println("Keyboard controls:")
	fmt.Println("  Z         - A button")
	fmt.Println("  X         - B button")
	fmt.Println("  Enter     - Start")
	fmt.Println("  Shift     - Select")
	fmt.Println("  Arrow keys - D-Pad")
	fmt.Println("  A         - L shoulder")
	fmt.Println("  S         - R shoulder")
	fmt.Println("")
	fmt.Println("You can also use File -> Open ROM from the menu")
}
