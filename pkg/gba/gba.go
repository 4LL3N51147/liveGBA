package gba

import (
	"fmt"
	"gba/pkg/apu"
	"gba/pkg/cartridge"
	"gba/pkg/cpu"
	"gba/pkg/dma"
	"gba/pkg/input"
	"gba/pkg/mmu"
	"gba/pkg/ppu"
	"gba/pkg/timer"
)

const (
	CPUFrequency   = 16777216
	CyclesPerFrame = 280896
)

type GBA struct {
	CPU   *cpu.CPU
	MMU   *mmu.MMU
	PPU   *ppu.PPU
	APU   *apu.APU
	DMA   *dma.DMA
	Timer *timer.Timer
	Input *input.Input

	Cartridge *cartridge.Cartridge

	FrameCount  int
	TotalCycles int64
}

func New() *GBA {
	gba := &GBA{}

	gba.MMU = mmu.New()
	gba.CPU = cpu.New()
	gba.PPU = ppu.New(gba.MMU.GetVRAM(), gba.MMU.GetPalette(), gba.MMU.GetOAM())
	gba.APU = apu.New()
	gba.DMA = dma.New(
		gba.MMU.Read32,
		gba.MMU.Write32,
		gba.MMU.Read16,
		gba.MMU.Write16,
		gba.RequestInterrupt,
	)
	gba.Timer = timer.New(gba.RequestInterrupt)
	gba.Input = input.New()

	gba.setupCallbacks()

	return gba
}

func (g *GBA) setupCallbacks() {
	g.CPU.Read8 = g.MMU.Read8
	g.CPU.Read16 = g.MMU.Read16
	g.CPU.Read32 = g.MMU.Read32
	g.CPU.Write8 = g.MMU.Write8
	g.CPU.Write16 = g.MMU.Write16
	g.CPU.Write32 = g.MMU.Write32
}

func (g *GBA) Reset() {
	g.CPU.Reset()
	g.MMU.Reset()
	g.PPU.Reset()
	g.APU.Reset()
	g.DMA.Reset()
	g.Timer.Reset()
	g.Input.Reset()

	// 检查是否有 BIOS 加载
	hasBIOS := len(g.MMU.GetBIOS()) > 0

	if hasBIOS {
		// 如果加载了 BIOS，让 BIOS 自己初始化寄存器
		// 只重置 PC 到 BIOS 开始位置
		g.CPU.PC = 0x00000000
		g.CPU.Regs[15] = g.CPU.PC + 4
		// 设置 R9 为 IWRAM 地址，帮助 BIOS 跳出循环
		// 真实 BIOS 会在初始化时设置 R9，但我们跳过初始化
		g.CPU.Regs[9] = 0x03000000
		fmt.Printf("[GBA] Reset: BIOS detected, starting from BIOS (0x00000000), R9=0x%08X\n", g.CPU.Regs[9])
	} else {
		// 没有 BIOS，使用默认初始化
		// R9 被 BIOS 设置为 IWRAM 起始地址，ROM 代码使用它作为基址
		g.CPU.Regs[9] = 0x03000000    // R9 = IWRAM 开始（BIOS 通常这样设置）
		g.CPU.Regs[10] = 0x00000000   // R10
		g.CPU.Regs[11] = 0x00000000   // R11
		g.CPU.Regs[12] = 0x00000000   // R12
		g.CPU.Regs[13] = 0x03007F00   // 用户模式 SP
		g.CPU.RegsSVC[0] = 0x03007FE0 // SVC 模式 SP
		g.CPU.RegsSVC[1] = 0x08000000 // SVC 模式 LR（返回地址）
		g.CPU.RegsIRQ[0] = 0x03007FA0 // IRQ 模式 SP

		// 设置 GBA 启动地址
		// 如果有 ROM，从 0x08000000 (ROM 开始) 执行
		// 如果没有 ROM，从 0x00000000 (BIOS) 执行
		if g.Cartridge != nil && len(g.Cartridge.GetROM()) > 0 {
			g.CPU.PC = 0x08000000
			g.CPU.Regs[15] = g.CPU.PC + 4
			fmt.Printf("[GBA] Reset: Set PC to ROM start (0x08000000), SP=0x%08X\n", g.CPU.Regs[13])
		} else {
			g.CPU.PC = 0x00000000
			g.CPU.Regs[15] = g.CPU.PC + 4
			fmt.Printf("[GBA] Reset: Set PC to BIOS start (0x00000000), SP=0x%08X\n", g.CPU.Regs[13])
		}
	}

	g.FrameCount = 0
	g.TotalCycles = 0
}

func (g *GBA) LoadROM(filename string) error {
	fmt.Printf("[GBA] Starting ROM load: %s\n", filename)

	cart, err := cartridge.Load(filename)
	if err != nil {
		fmt.Printf("[GBA] ERROR: Failed to load cartridge: %v\n", err)
		return fmt.Errorf("failed to load cartridge: %w", err)
	}

	g.Cartridge = cart
	fmt.Printf("[GBA] Cartridge loaded, ROM size: %d bytes\n", len(cart.GetROM()))

	g.MMU.LoadROM(cart.GetROM())
	fmt.Printf("[GBA] ROM loaded into MMU\n")

	g.Reset()
	fmt.Printf("[GBA] System reset complete\n")
	fmt.Printf("[GBA] Initial PC: 0x%08X, CPSR: 0x%08X\n", g.CPU.PC, g.CPU.CPSR)

	return nil
}

func (g *GBA) LoadBIOS(filename string) error {
	data, err := cartridge.Load(filename)
	if err != nil {
		return fmt.Errorf("failed to load BIOS: %w", err)
	}

	if len(data.GetROM()) != 0x4000 {
		return fmt.Errorf("BIOS must be exactly 16KB")
	}

	g.MMU.LoadBIOS(data.GetROM())
	return nil
}

func (g *GBA) Step() int {
	cycles := g.CPU.Step()

	g.TotalCycles += int64(cycles)

	g.Timer.Step(cycles)
	g.APU.Step(cycles)
	g.DMA.Step()

	if g.PPU.Step(cycles) {
		g.FrameCount++
		if g.MMU.CheckInterrupts() {
			g.handleInterrupt()
		}
	}

	if g.MMU.CheckInterrupts() {
		g.handleInterrupt()
	}

	return cycles
}

func (g *GBA) RunFrame() {
	cycles := 0
	frameStart := g.FrameCount

	for cycles < CyclesPerFrame {
		cycles += g.Step()
	}

	// 每 60 帧（约 1 秒）输出一次日志
	if g.FrameCount%60 == 0 && g.FrameCount != frameStart {
		fmt.Printf("[GBA] Frame %d, PC: 0x%08X, Total Cycles: %d\n",
			g.FrameCount, g.CPU.PC, g.TotalCycles)
	}
}

func (g *GBA) handleInterrupt() {
	if g.CPU.GetFlag(cpu.FlagI) {
		return
	}

	if g.MMU.IME == 0 {
		return
	}

	irq := g.MMU.IE & g.MMU.IF
	if irq == 0 {
		return
	}

	fmt.Printf("[GBA] INTERRUPT! PC=0x%08X -> 0x00000018, IRQ=0x%04X\n", g.CPU.PC, irq)

	g.CPU.SaveMode()
	g.CPU.SwitchMode(cpu.ModeIRQ)
	g.CPU.SetSPSR(g.CPU.CPSR)
	g.CPU.CPSR |= cpu.FlagI | cpu.FlagT
	g.CPU.LR = g.CPU.PC - 4
	g.CPU.PC = 0x00000018
	g.CPU.Regs[15] = g.CPU.PC + 4
}

func (g *GBA) RequestInterrupt(irq uint16) {
	g.MMU.RequestInterrupt(irq)
}

func (g *GBA) GetFrameBuffer() []uint16 {
	return g.PPU.GetFrameBuffer()
}

func (g *GBA) SetKey(key int, pressed bool) {
	g.Input.SetKey(key, pressed)
	g.MMU.KEYINPUT = g.Input.GetKEYINPUT()
}

func (g *GBA) GetFPS() float64 {
	return float64(g.FrameCount) / (float64(g.TotalCycles) / CPUFrequency)
}

func (g *GBA) GetFrameCount() int {
	return g.FrameCount
}
