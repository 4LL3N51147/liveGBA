package cartridge

import (
	"encoding/binary"
	"fmt"
	"os"
)

type Cartridge struct {
	ROM       []byte
	Title     string
	GameCode  string
	MakerCode string
	UnitCode  byte
	Version   byte
	Checksum  byte
}

func Load(filename string) (*Cartridge, error) {
	fmt.Printf("[Cartridge] Loading ROM file: %s\n", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("[Cartridge] ERROR: Failed to read ROM file: %v\n", err)
		return nil, fmt.Errorf("failed to read ROM file: %w", err)
	}

	fmt.Printf("[Cartridge] Read %d bytes from file\n", len(data))

	if len(data) < 0xC0 {
		fmt.Printf("[Cartridge] ERROR: ROM file too small (%d bytes)\n", len(data))
		return nil, fmt.Errorf("ROM file too small")
	}

	cart := &Cartridge{
		ROM: data,
	}

	cart.parseHeader()
	fmt.Printf("[Cartridge] ROM Info: %s\n", cart.String())

	return cart, nil
}

func (c *Cartridge) parseHeader() {
	if len(c.ROM) >= 0xA0 {
		c.Title = string(c.ROM[0xA0:0xAC])
	}
	if len(c.ROM) >= 0xAC {
		c.GameCode = string(c.ROM[0xAC:0xB0])
	}
	if len(c.ROM) >= 0xB0 {
		c.MakerCode = string(c.ROM[0xB0:0xB2])
	}
	if len(c.ROM) >= 0xB2 {
		c.UnitCode = c.ROM[0xB2]
	}
	if len(c.ROM) >= 0xBC {
		c.Version = c.ROM[0xBC]
	}
	if len(c.ROM) >= 0xBD {
		c.Checksum = c.ROM[0xBD]
	}
}

func (c *Cartridge) GetROM() []byte {
	return c.ROM
}

func (c *Cartridge) GetSize() int {
	return len(c.ROM)
}

func (c *Cartridge) Read8(addr uint32) uint8 {
	if addr < uint32(len(c.ROM)) {
		return c.ROM[addr]
	}
	return 0
}

func (c *Cartridge) Read16(addr uint32) uint16 {
	if addr+1 < uint32(len(c.ROM)) {
		return binary.LittleEndian.Uint16(c.ROM[addr:])
	}
	return 0
}

func (c *Cartridge) Read32(addr uint32) uint32 {
	if addr+3 < uint32(len(c.ROM)) {
		return binary.LittleEndian.Uint32(c.ROM[addr:])
	}
	return 0
}

func (c *Cartridge) GetSaveType() string {
	if len(c.ROM) < 0xE4 {
		return "SRAM"
	}

	id := binary.LittleEndian.Uint32(c.ROM[0xAC:])

	switch id {
	case 0x45564153:
		return "SRAM"
	case 0x52414D5F:
		return "EEPROM"
	case 0x53414D5F:
		return "SRAM"
	case 0x53414D46:
		return "FLASH"
	default:
		return "SRAM"
	}
}

func (c *Cartridge) String() string {
	return fmt.Sprintf("Title: %s, GameCode: %s, Maker: %s, Size: %d bytes",
		c.Title, c.GameCode, c.MakerCode, len(c.ROM))
}
