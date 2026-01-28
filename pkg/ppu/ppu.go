package ppu

import (
	"fmt"
	"image"
	"image/color"
)

var _ = color.RGBA{}

const (
	ScreenWidth  = 240
	ScreenHeight = 160

	Mode0 = 0
	Mode1 = 1
	Mode2 = 2
	Mode3 = 3
	Mode4 = 4
	Mode5 = 5

	BGModeMask   = 0x0007
	FrameSelect  = 0x0010
	HBlankFree   = 0x0020
	Obj1DMap     = 0x0040
	ForcedBlank  = 0x0080
	BG0Enable    = 0x0100
	BG1Enable    = 0x0200
	BG2Enable    = 0x0400
	BG3Enable    = 0x0800
	ObjEnable    = 0x1000
	Win0Enable   = 0x2000
	Win1Enable   = 0x4000
	ObjWinEnable = 0x8000

	VBlankFlag = 0x0001
	HBlankFlag = 0x0002
	VCountFlag = 0x0004
	VBlankIRQ  = 0x0008
	HBlankIRQ  = 0x0010
	VCountIRQ  = 0x0020
)

type PPU struct {
	VRAM    []byte
	Palette []byte
	OAM     []byte

	DISPCNT  uint16
	DISPSTAT uint16
	VCOUNT   uint16

	BG0CNT  uint16
	BG1CNT  uint16
	BG2CNT  uint16
	BG3CNT  uint16
	BG0HOFS uint16
	BG0VOFS uint16
	BG1HOFS uint16
	BG1VOFS uint16
	BG2HOFS uint16
	BG2VOFS uint16
	BG3HOFS uint16
	BG3VOFS uint16

	BG2PA   int16
	BG2PB   int16
	BG2PC   int16
	BG2PD   int16
	BG2X    int32
	BG2Y    int32
	BG2RefX int32
	BG2RefY int32

	BG3PA   int16
	BG3PB   int16
	BG3PC   int16
	BG3PD   int16
	BG3X    int32
	BG3Y    int32
	BG3RefX int32
	BG3RefY int32

	WIN0H    uint16
	WIN1H    uint16
	WIN0V    uint16
	WIN1V    uint16
	WININ    uint16
	WINOUT   uint16
	MOSAIC   uint16
	BLDCNT   uint16
	BLDALPHA uint16
	BLDY     uint16

	FrameBuffer []uint16
	CurrentLine int
	CycleCount  int

	HBlank bool
	VBlank bool
}

func New(vram, palette, oam []byte) *PPU {
	ppu := &PPU{
		VRAM:        vram,
		Palette:     palette,
		OAM:         oam,
		FrameBuffer: make([]uint16, ScreenWidth*ScreenHeight),
	}
	ppu.Reset()
	return ppu
}

func (p *PPU) Reset() {
	p.DISPCNT = 0x0080
	p.DISPSTAT = 0x0000
	p.VCOUNT = 0x0000
	p.CurrentLine = 0
	p.CycleCount = 0
	p.HBlank = false
	p.VBlank = false

	for i := range p.FrameBuffer {
		p.FrameBuffer[i] = 0
	}

	fmt.Printf("[PPU] Reset complete. DISPCNT: 0x%04X (ForcedBlank: %v)\n",
		p.DISPCNT, p.DISPCNT&ForcedBlank != 0)
}

func (p *PPU) Step(cycles int) bool {
	p.CycleCount += cycles

	if p.VBlank {
		if p.CycleCount >= 1232 {
			p.CycleCount -= 1232
			p.CurrentLine++
			p.VCOUNT = uint16(p.CurrentLine)

			if p.CurrentLine >= 228 {
				p.CurrentLine = 0
				p.VCOUNT = 0
				p.VBlank = false
				p.DISPSTAT &^= VBlankFlag
				return true
			}
		}
	} else if p.HBlank {
		if p.CycleCount >= 272 {
			p.CycleCount -= 272
			p.HBlank = false
			p.DISPSTAT &^= HBlankFlag
			p.CurrentLine++
			p.VCOUNT = uint16(p.CurrentLine)

			if p.CurrentLine >= 160 {
				p.VBlank = true
				p.DISPSTAT |= VBlankFlag
				if p.DISPSTAT&VBlankIRQ != 0 {
					return true
				}
			}
		}
	} else {
		if p.CycleCount >= 960 {
			p.CycleCount -= 960
			p.HBlank = true
			p.DISPSTAT |= HBlankFlag

			if p.DISPSTAT&HBlankIRQ != 0 {
				return true
			}

			if !p.VBlank {
				p.RenderScanline()
			}
		}
	}

	return false
}

func (p *PPU) RenderScanline() {
	mode := p.DISPCNT & BGModeMask

	if p.DISPCNT&ForcedBlank != 0 {
		// 强制空白 - 填充白色
		for x := 0; x < ScreenWidth; x++ {
			p.FrameBuffer[p.CurrentLine*ScreenWidth+x] = 0x7FFF
		}
		return
	}

	switch mode {
	case Mode0:
		p.renderMode0()
	case Mode1:
		p.renderMode1()
	case Mode2:
		p.renderMode2()
	case Mode3:
		p.renderMode3()
	case Mode4:
		p.renderMode4()
	case Mode5:
		p.renderMode5()
	}
}

func (p *PPU) renderMode0() {
	for x := 0; x < ScreenWidth; x++ {
		p.FrameBuffer[p.CurrentLine*ScreenWidth+x] = p.getPaletteColor(0)
	}
}

func (p *PPU) renderMode1() {
	for x := 0; x < ScreenWidth; x++ {
		p.FrameBuffer[p.CurrentLine*ScreenWidth+x] = p.getPaletteColor(0)
	}
}

func (p *PPU) renderMode2() {
	for x := 0; x < ScreenWidth; x++ {
		p.FrameBuffer[p.CurrentLine*ScreenWidth+x] = p.getPaletteColor(0)
	}
}

func (p *PPU) renderMode3() {
	if p.DISPCNT&FrameSelect != 0 {
		return
	}

	for x := 0; x < ScreenWidth; x++ {
		offset := (p.CurrentLine*ScreenWidth + x) * 2
		if offset+1 < len(p.VRAM) {
			color := uint16(p.VRAM[offset]) | uint16(p.VRAM[offset+1])<<8
			p.FrameBuffer[p.CurrentLine*ScreenWidth+x] = color
		}
	}
}

func (p *PPU) renderMode4() {
	frameOffset := 0
	if p.DISPCNT&FrameSelect != 0 {
		frameOffset = 0xA000
	}

	for x := 0; x < ScreenWidth; x++ {
		offset := frameOffset + p.CurrentLine*ScreenWidth + x
		if offset < len(p.VRAM) {
			paletteIdx := p.VRAM[offset]
			p.FrameBuffer[p.CurrentLine*ScreenWidth+x] = p.getPaletteColor(paletteIdx)
		}
	}
}

func (p *PPU) renderMode5() {
	if p.CurrentLine >= 128 {
		for x := 0; x < ScreenWidth; x++ {
			p.FrameBuffer[p.CurrentLine*ScreenWidth+x] = 0
		}
		return
	}

	frameOffset := 0
	if p.DISPCNT&FrameSelect != 0 {
		frameOffset = 0xA000
	}

	for x := 0; x < 160; x++ {
		offset := frameOffset + (p.CurrentLine*160+x)*2
		if offset+1 < len(p.VRAM) {
			color := uint16(p.VRAM[offset]) | uint16(p.VRAM[offset+1])<<8
			if x < ScreenWidth {
				p.FrameBuffer[p.CurrentLine*ScreenWidth+x] = color
			}
		}
	}

	for x := 160; x < ScreenWidth; x++ {
		p.FrameBuffer[p.CurrentLine*ScreenWidth+x] = 0
	}
}

func (p *PPU) getPaletteColor(idx uint8) uint16 {
	offset := uint32(idx) * 2
	if offset+1 < uint32(len(p.Palette)) {
		return uint16(p.Palette[offset]) | uint16(p.Palette[offset+1])<<8
	}
	return 0
}

func (p *PPU) GetFrameBuffer() []uint16 {
	return p.FrameBuffer
}

func (p *PPU) ToImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, ScreenWidth, ScreenHeight))

	for y := 0; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			col := p.FrameBuffer[y*ScreenWidth+x]
			r := uint8((col & 0x1F) << 3)
			g := uint8(((col >> 5) & 0x1F) << 3)
			b := uint8(((col >> 10) & 0x1F) << 3)
			img.SetRGBA(x, y, color.RGBA{r, g, b, 255})
		}
	}

	return img
}

func (p *PPU) SetDISPCNT(val uint16) {
	oldVal := p.DISPCNT
	p.DISPCNT = val

	// 检测从强制空白切换到正常显示
	if oldVal&ForcedBlank != 0 && val&ForcedBlank == 0 {
		fmt.Printf("[PPU] DISPCNT changed: 0x%04X -> 0x%04X (ForcedBlank OFF, Mode: %d)\n",
			oldVal, val, val&BGModeMask)
	}
}

func (p *PPU) SetDISPSTAT(val uint16) {
	p.DISPSTAT = (p.DISPSTAT & 0x0007) | (val & 0xFFF8)
}

func (p *PPU) GetVCOUNT() uint16 {
	return p.VCOUNT
}

func (p *PPU) GetDISPSTAT() uint16 {
	return p.DISPSTAT
}

func (p *PPU) IsVBlank() bool {
	return p.VBlank
}

func (p *PPU) IsHBlank() bool {
	return p.HBlank
}

func (p *PPU) ReadRegister(addr uint32) uint16 {
	switch addr {
	case 0x04000000:
		return p.DISPCNT
	case 0x04000004:
		return p.DISPSTAT
	case 0x04000006:
		return p.VCOUNT
	case 0x04000008:
		return p.BG0CNT
	case 0x0400000A:
		return p.BG1CNT
	case 0x0400000C:
		return p.BG2CNT
	case 0x0400000E:
		return p.BG3CNT
	default:
		return 0
	}
}

func (p *PPU) WriteRegister(addr uint32, val uint16) {
	switch addr {
	case 0x04000000:
		p.DISPCNT = val
	case 0x04000004:
		p.SetDISPSTAT(val)
	case 0x04000008:
		p.BG0CNT = val
	case 0x0400000A:
		p.BG1CNT = val
	case 0x0400000C:
		p.BG2CNT = val
	case 0x0400000E:
		p.BG3CNT = val
	case 0x04000010:
		p.BG0HOFS = val
	case 0x04000012:
		p.BG0VOFS = val
	case 0x04000014:
		p.BG1HOFS = val
	case 0x04000016:
		p.BG1VOFS = val
	case 0x04000018:
		p.BG2HOFS = val
	case 0x0400001A:
		p.BG2VOFS = val
	case 0x0400001C:
		p.BG3HOFS = val
	case 0x0400001E:
		p.BG3VOFS = val
	}
}
