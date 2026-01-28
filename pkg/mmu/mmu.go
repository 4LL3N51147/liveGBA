package mmu

import (
	"encoding/binary"
	"fmt"
)

const (
	BIOSStart  = 0x00000000
	BIOSLength = 0x00004000

	WRAM256Start  = 0x02000000
	WRAM256Length = 0x00040000

	WRAM32Start  = 0x03000000
	WRAM32Length = 0x00008000

	IOStart  = 0x04000000
	IOLength = 0x000003FF

	PaletteStart  = 0x05000000
	PaletteLength = 0x00000400

	VRAMStart = 0x06000000
	VRAMLenth = 0x00018000

	OAMStart  = 0x07000000
	OAMLength = 0x00000400

	ROMStart  = 0x08000000
	ROMLength = 0x02000000

	SRAMStart = 0x0E000000
	SRAMLenth = 0x00010000
)

type MMU struct {
	BIOS    []byte
	WRAM256 []byte
	WRAM32  []byte
	IO      []byte
	Palette []byte
	VRAM    []byte
	OAM     []byte
	ROM     []byte
	SRAM    []byte

	WaitStates [4]int

	DISPCNT  uint16
	DISPSTAT uint16
	VCOUNT   uint16
	BG0CNT   uint16
	BG1CNT   uint16
	BG2CNT   uint16
	BG3CNT   uint16
	BG0HOFS  uint16
	BG0VOFS  uint16
	BG1HOFS  uint16
	BG1VOFS  uint16
	BG2HOFS  uint16
	BG2VOFS  uint16
	BG3HOFS  uint16
	BG3VOFS  uint16
	BG2PA    uint16
	BG2PB    uint16
	BG2PC    uint16
	BG2PD    uint16
	BG2X     uint32
	BG2Y     uint32
	BG3PA    uint16
	BG3PB    uint16
	BG3PC    uint16
	BG3PD    uint16
	BG3X     uint32
	BG3Y     uint32
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

	SOUND1CNT_L uint16
	SOUND1CNT_H uint16
	SOUND1CNT_X uint16
	SOUND2CNT_L uint16
	SOUND2CNT_H uint16
	SOUND3CNT_L uint16
	SOUND3CNT_H uint16
	SOUND3CNT_X uint16
	SOUND4CNT_L uint16
	SOUND4CNT_H uint16
	SOUNDCNT_L  uint16
	SOUNDCNT_H  uint16
	SOUNDCNT_X  uint16
	SOUNDBIAS   uint16
	WAVE_RAM    [16]byte
	FIFO_A      [4]byte
	FIFO_B      [4]byte

	DMA0SAD   uint32
	DMA0DAD   uint32
	DMA0CNT_L uint16
	DMA0CNT_H uint16
	DMA1SAD   uint32
	DMA1DAD   uint32
	DMA1CNT_L uint16
	DMA1CNT_H uint16
	DMA2SAD   uint32
	DMA2DAD   uint32
	DMA2CNT_L uint16
	DMA2CNT_H uint16
	DMA3SAD   uint32
	DMA3DAD   uint32
	DMA3CNT_L uint16
	DMA3CNT_H uint16

	TM0CNT_L uint16
	TM0CNT_H uint16
	TM1CNT_L uint16
	TM1CNT_H uint16
	TM2CNT_L uint16
	TM2CNT_H uint16
	TM3CNT_L uint16
	TM3CNT_H uint16

	SIODATA32   uint32
	SIOMULTI0   uint16
	SIOMULTI1   uint16
	SIOMULTI2   uint16
	SIOMULTI3   uint16
	SIOCNT      uint16
	SIOMLT_SEND uint16

	KEYINPUT uint16
	KEYCNT   uint16

	RCNT uint16
	IR   uint16

	JOYCNT    uint16
	JOY_RECV  uint32
	JOY_TRANS uint32
	JOYSTAT   uint16

	IE      uint16
	IF      uint16
	WAITCNT uint16
	IME     uint16

	POSTFLG uint8
	HALTCNT uint8

	InternalRAM []byte
}

func New() *MMU {
	mmu := &MMU{
		BIOS:       make([]byte, BIOSLength),
		WRAM256:    make([]byte, WRAM256Length),
		WRAM32:     make([]byte, WRAM32Length),
		IO:         make([]byte, IOLength),
		Palette:    make([]byte, PaletteLength),
		VRAM:       make([]byte, VRAMLenth),
		OAM:        make([]byte, OAMLength),
		ROM:        make([]byte, 0),
		SRAM:       make([]byte, SRAMLenth),
		WaitStates: [4]int{4, 3, 2, 8},
	}
	mmu.Reset()
	return mmu
}

func (m *MMU) Reset() {
	for i := range m.WRAM256 {
		m.WRAM256[i] = 0
	}
	for i := range m.WRAM32 {
		m.WRAM32[i] = 0
	}
	for i := range m.IO {
		m.IO[i] = 0
	}
	for i := range m.Palette {
		m.Palette[i] = 0
	}
	for i := range m.VRAM {
		m.VRAM[i] = 0
	}
	for i := range m.OAM {
		m.OAM[i] = 0
	}
	for i := range m.SRAM {
		m.SRAM[i] = 0
	}

	m.DISPCNT = 0x0080
	m.DISPSTAT = 0x0000
	m.VCOUNT = 0x0000
	m.IE = 0x0000
	m.IF = 0x0000
	m.IME = 0x0000
	m.WAITCNT = 0x0000
	m.KEYINPUT = 0x03FF
}

func (m *MMU) LoadBIOS(data []byte) {
	copy(m.BIOS, data)
}

func (m *MMU) LoadROM(data []byte) {
	m.ROM = make([]byte, len(data))
	copy(m.ROM, data)
}

func (m *MMU) mirrorAddress(addr uint32) uint32 {
	switch {
	case addr >= 0x00000000 && addr < 0x00004000:
		return addr
	case addr >= 0x02000000 && addr < 0x02040000:
		return addr
	case addr >= 0x03000000 && addr < 0x03008000:
		return addr
	case addr >= 0x04000000 && addr < 0x04000400:
		return addr
	case addr >= 0x05000000 && addr < 0x05000400:
		return addr
	case addr >= 0x06000000 && addr < 0x06018000:
		if addr >= 0x06010000 {
			return addr - 0x8000
		}
		return addr
	case addr >= 0x07000000 && addr < 0x07000400:
		return addr
	case addr >= 0x08000000 && addr < 0x0E000000:
		return addr
	case addr >= 0x0E000000:
		return addr
	default:
		return addr
	}
}

func (m *MMU) Read8(addr uint32) uint8 {
	addr = m.mirrorAddress(addr)

	switch {
	case addr >= BIOSStart && addr < BIOSStart+BIOSLength:
		if addr == 0x18 || addr == 0x19 || addr == 0x1A || addr == 0x1B {
			fmt.Printf("[MMU] BIOS read at 0x%08X = 0x%02X\n", addr, m.BIOS[addr])
		}
		return m.BIOS[addr]
	case addr >= WRAM256Start && addr < WRAM256Start+WRAM256Length:
		return m.WRAM256[addr-WRAM256Start]
	case addr >= WRAM32Start && addr < WRAM32Start+WRAM32Length:
		return m.WRAM32[addr-WRAM32Start]
	case addr >= IOStart && addr < IOStart+IOLength:
		return m.readIO8(addr)
	case addr >= PaletteStart && addr < PaletteStart+PaletteLength:
		return m.Palette[addr-PaletteStart]
	case addr >= VRAMStart && addr < VRAMStart+VRAMLenth:
		return m.VRAM[addr-VRAMStart]
	case addr >= OAMStart && addr < OAMStart+OAMLength:
		return m.OAM[addr-OAMStart]
	case addr >= ROMStart && addr < ROMStart+uint32(len(m.ROM)):
		return m.ROM[addr-ROMStart]
	case addr >= SRAMStart && addr < SRAMStart+SRAMLenth:
		return m.SRAM[addr-SRAMStart]
	default:
		return 0
	}
}

func (m *MMU) Read16(addr uint32) uint16 {
	return uint16(m.Read8(addr)) | uint16(m.Read8(addr+1))<<8
}

func (m *MMU) Read32(addr uint32) uint32 {
	return uint32(m.Read16(addr)) | uint32(m.Read16(addr+2))<<16
}

func (m *MMU) Write8(addr uint32, val uint8) {
	addr = m.mirrorAddress(addr)

	switch {
	case addr >= BIOSStart && addr < BIOSStart+BIOSLength:
		// BIOS is read-only
	case addr >= WRAM256Start && addr < WRAM256Start+WRAM256Length:
		m.WRAM256[addr-WRAM256Start] = val
	case addr >= WRAM32Start && addr < WRAM32Start+WRAM32Length:
		m.WRAM32[addr-WRAM32Start] = val
	case addr >= IOStart && addr < IOStart+IOLength:
		m.writeIO8(addr, val)
	case addr >= PaletteStart && addr < PaletteStart+PaletteLength:
		m.Palette[addr-PaletteStart] = val
	case addr >= VRAMStart && addr < VRAMStart+VRAMLenth:
		m.VRAM[addr-VRAMStart] = val
	case addr >= OAMStart && addr < OAMStart+OAMLength:
		m.OAM[addr-OAMStart] = val
	case addr >= ROMStart:
		// ROM is read-only
	case addr >= SRAMStart:
		m.SRAM[addr-SRAMStart] = val
	}
}

func (m *MMU) Write16(addr uint32, val uint16) {
	m.Write8(addr, uint8(val))
	m.Write8(addr+1, uint8(val>>8))
}

func (m *MMU) Write32(addr uint32, val uint32) {
	m.Write16(addr, uint16(val))
	m.Write16(addr+2, uint16(val>>16))
}

func (m *MMU) readIO8(addr uint32) uint8 {
	offset := addr - IOStart

	switch offset {
	case 0x00, 0x01:
		return uint8(m.DISPCNT >> ((offset & 1) * 8))
	case 0x04, 0x05:
		return uint8(m.DISPSTAT >> ((offset & 1) * 8))
	case 0x06, 0x07:
		return uint8(m.VCOUNT >> ((offset & 1) * 8))
	case 0x130, 0x131:
		return uint8(m.KEYINPUT >> ((offset & 1) * 8))
	case 0x200, 0x201:
		return uint8(m.IE >> ((offset & 1) * 8))
	case 0x202, 0x203:
		return uint8(m.IF >> ((offset & 1) * 8))
	case 0x208, 0x209:
		return uint8(m.IME >> ((offset & 1) * 8))
	default:
		if int(offset) < len(m.IO) {
			return m.IO[offset]
		}
		return 0
	}
}

func (m *MMU) writeIO8(addr uint32, val uint8) {
	offset := addr - IOStart

	switch offset {
	case 0x00:
		oldVal := m.DISPCNT
		m.DISPCNT = (m.DISPCNT & 0xFF00) | uint16(val)
		if oldVal != m.DISPCNT {
			fmt.Printf("[MMU] DISPCNT write: 0x%04X -> 0x%04X (ForcedBlank: %v)\n",
				oldVal, m.DISPCNT, m.DISPCNT&0x0080 != 0)
		}
	case 0x01:
		oldVal := m.DISPCNT
		m.DISPCNT = (m.DISPCNT & 0x00FF) | (uint16(val) << 8)
		if oldVal != m.DISPCNT {
			fmt.Printf("[MMU] DISPCNT write: 0x%04X -> 0x%04X (Mode: %d, ForcedBlank: %v)\n",
				oldVal, m.DISPCNT, m.DISPCNT&0x0007, m.DISPCNT&0x0080 != 0)
		}
	case 0x04:
		m.DISPSTAT = (m.DISPSTAT & 0xFF00) | uint16(val)
	case 0x05:
		m.DISPSTAT = (m.DISPSTAT & 0x00FF) | (uint16(val) << 8)
	case 0x130, 0x131:
		// KEYINPUT is read-only
	case 0x200:
		m.IE = (m.IE & 0xFF00) | uint16(val)
	case 0x201:
		m.IE = (m.IE & 0x00FF) | (uint16(val) << 8)
	case 0x202:
		m.IF &= ^uint16(val)
	case 0x203:
		m.IF &= ^(uint16(val) << 8)
	case 0x208:
		m.IME = uint16(val) & 1
	case 0x301:
		m.HALTCNT = val
	default:
		if int(offset) < len(m.IO) {
			m.IO[offset] = val
		}
	}
}

func (m *MMU) GetVRAM() []byte {
	return m.VRAM
}

func (m *MMU) GetPalette() []byte {
	return m.Palette
}

func (m *MMU) GetOAM() []byte {
	return m.OAM
}

func (m *MMU) RequestInterrupt(irq uint16) {
	m.IF |= irq
}

func (m *MMU) CheckInterrupts() bool {
	return m.IME != 0 && (m.IE&m.IF) != 0
}

func (m *MMU) GetBIOS() []byte {
	return m.BIOS
}

func (m *MMU) GetROM() []byte {
	return m.ROM
}

func (m *MMU) LoadSave(data []byte) {
	copy(m.SRAM, data)
}

func (m *MMU) GetSaveData() []byte {
	data := make([]byte, len(m.SRAM))
	copy(data, m.SRAM)
	return data
}

func (m *MMU) ReadBIOS(addr uint32) uint32 {
	if addr >= BIOSStart && addr+3 < BIOSStart+BIOSLength {
		return binary.LittleEndian.Uint32(m.BIOS[addr:])
	}
	return 0
}
