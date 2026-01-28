# GBA Emulator in Go

A Game Boy Advance emulator written in Go.

## Features

- ARM7TDMI CPU emulation
- Memory management
- Graphics rendering
- Audio support
- Input handling

## Usage

```bash
go run cmd/gba/main.go <rom_file.gba>
```

## Project Structure

- `pkg/cpu` - ARM7TDMI CPU implementation
- `pkg/mmu` - Memory Management Unit
- `pkg/ppu` - Picture Processing Unit (graphics)
- `pkg/apu` - Audio Processing Unit
- `pkg/dma` - Direct Memory Access controller
- `pkg/timer` - Timer system
- `pkg/input` - Input handling
- `pkg/cartridge` - ROM cartridge handling
- `cmd/gba` - Main application
