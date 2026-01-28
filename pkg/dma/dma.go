package dma

const (
	DMA0CNT_H = 0x040000BA
	DMA1CNT_H = 0x040000C6
	DMA2CNT_H = 0x040000D2
	DMA3CNT_H = 0x040000DE
)

type DMA struct {
	SAD   [4]uint32
	DAD   [4]uint32
	CNT_L [4]uint16
	CNT_H [4]uint16

	Active         [4]bool
	InternalSource [4]uint32
	InternalDest   [4]uint32
	InternalCount  [4]uint32

	Read32           func(addr uint32) uint32
	Write32          func(addr uint32, val uint32)
	Read16           func(addr uint32) uint16
	Write16          func(addr uint32, val uint16)
	RequestInterrupt func(irq uint16)
}

func New(read32 func(uint32) uint32, write32 func(uint32, uint32),
	read16 func(uint32) uint16, write16 func(uint32, uint16),
	requestIRQ func(uint16)) *DMA {
	dma := &DMA{
		Read32:           read32,
		Write32:          write32,
		Read16:           read16,
		Write16:          write16,
		RequestInterrupt: requestIRQ,
	}
	dma.Reset()
	return dma
}

func (d *DMA) Reset() {
	for i := 0; i < 4; i++ {
		d.SAD[i] = 0
		d.DAD[i] = 0
		d.CNT_L[i] = 0
		d.CNT_H[i] = 0
		d.Active[i] = false
		d.InternalSource[i] = 0
		d.InternalDest[i] = 0
		d.InternalCount[i] = 0
	}
}

func (d *DMA) WriteSAD(channel int, val uint32) {
	if channel < 4 {
		d.SAD[channel] = val & 0x0FFFFFFF
	}
}

func (d *DMA) WriteDAD(channel int, val uint32) {
	if channel < 4 {
		d.DAD[channel] = val & 0x0FFFFFFF
	}
}

func (d *DMA) WriteCNT_L(channel int, val uint16) {
	if channel < 4 {
		d.CNT_L[channel] = val
	}
}

func (d *DMA) WriteCNT_H(channel int, val uint16) {
	if channel >= 4 {
		return
	}

	oldEnable := d.CNT_H[channel]&0x8000 != 0
	d.CNT_H[channel] = val

	if !oldEnable && val&0x8000 != 0 {
		d.startTransfer(channel)
	}
}

func (d *DMA) startTransfer(channel int) {
	d.InternalSource[channel] = d.SAD[channel]
	d.InternalDest[channel] = d.DAD[channel]

	count := uint32(d.CNT_L[channel])
	if count == 0 {
		if channel == 3 {
			count = 0x10000
		} else {
			count = 0x4000
		}
	}
	d.InternalCount[channel] = count

	startTiming := (d.CNT_H[channel] >> 12) & 0x3

	if startTiming == 0 {
		d.Active[channel] = true
	}
}

func (d *DMA) Step() int {
	cycles := 0

	for i := 0; i < 4; i++ {
		if d.Active[i] {
			cycles += d.transfer(i)
		}
	}

	return cycles
}

func (d *DMA) transfer(channel int) int {
	transferType := (d.CNT_H[channel] >> 10) & 0x3
	sourceControl := (d.CNT_H[channel] >> 7) & 0x3
	destControl := (d.CNT_H[channel] >> 5) & 0x3

	words := d.InternalCount[channel]
	if words > 16 {
		words = 16
	}

	for i := uint32(0); i < words; i++ {
		if transferType == 3 {
			data := d.Read32(d.InternalSource[channel])
			d.Write32(d.InternalDest[channel], data)
		} else if transferType == 2 {
			data := d.Read32(d.InternalSource[channel])
			d.Write32(d.InternalDest[channel], data)
		} else if transferType == 1 {
			data := d.Read16(d.InternalSource[channel])
			d.Write16(d.InternalDest[channel], data)
		} else {
			data := d.Read16(d.InternalSource[channel])
			d.Write16(d.InternalDest[channel], data)
		}

		d.InternalSource[channel] = d.adjustAddress(d.InternalSource[channel], sourceControl, transferType)
		d.InternalDest[channel] = d.adjustAddress(d.InternalDest[channel], destControl, transferType)
	}

	d.InternalCount[channel] -= words

	if d.InternalCount[channel] == 0 {
		d.Active[channel] = false

		if d.CNT_H[channel]&0x4000 != 0 {
			irq := uint16(1 << (8 + channel))
			d.RequestInterrupt(irq)
		}

		if d.CNT_H[channel]&0x0200 != 0 {
			d.startTransfer(channel)
		}
	}

	return int(words) * 2
}

func (d *DMA) adjustAddress(addr uint32, control uint16, transferType uint16) uint32 {
	increment := uint32(2)
	if transferType == 2 || transferType == 3 {
		increment = 4
	}

	switch control {
	case 0:
		return addr + increment
	case 1:
		return addr - increment
	case 2:
		return addr
	case 3:
		return addr + increment
	}
	return addr
}

func (d *DMA) Trigger(startTiming int) {
	for i := 0; i < 4; i++ {
		timing := int((d.CNT_H[i] >> 12) & 0x3)
		if timing == startTiming && d.CNT_H[i]&0x8000 != 0 {
			d.startTransfer(i)
			d.Active[i] = true
		}
	}
}

func (d *DMA) ReadRegister(addr uint32) uint16 {
	switch addr {
	case 0x040000B0, 0x040000B2:
		return uint16(d.SAD[0] >> ((addr & 2) * 8))
	case 0x040000B4, 0x040000B6:
		return uint16(d.DAD[0] >> ((addr & 2) * 8))
	case 0x040000B8:
		return d.CNT_L[0]
	case 0x040000BA:
		return d.CNT_H[0]
	case 0x040000BC, 0x040000BE:
		return uint16(d.SAD[1] >> ((addr & 2) * 8))
	case 0x040000C0, 0x040000C2:
		return uint16(d.DAD[1] >> ((addr & 2) * 8))
	case 0x040000C4:
		return d.CNT_L[1]
	case 0x040000C6:
		return d.CNT_H[1]
	case 0x040000C8, 0x040000CA:
		return uint16(d.SAD[2] >> ((addr & 2) * 8))
	case 0x040000CC, 0x040000CE:
		return uint16(d.DAD[2] >> ((addr & 2) * 8))
	case 0x040000D0:
		return d.CNT_L[2]
	case 0x040000D2:
		return d.CNT_H[2]
	case 0x040000D4, 0x040000D6:
		return uint16(d.SAD[3] >> ((addr & 2) * 8))
	case 0x040000D8, 0x040000DA:
		return uint16(d.DAD[3] >> ((addr & 2) * 8))
	case 0x040000DC:
		return d.CNT_L[3]
	case 0x040000DE:
		return d.CNT_H[3]
	}
	return 0
}

func (d *DMA) WriteRegister(addr uint32, val uint16) {
	switch addr {
	case 0x040000B0:
		d.SAD[0] = (d.SAD[0] & 0xFFFF0000) | uint32(val)
	case 0x040000B2:
		d.SAD[0] = (d.SAD[0] & 0x0000FFFF) | (uint32(val) << 16)
	case 0x040000B4:
		d.DAD[0] = (d.DAD[0] & 0xFFFF0000) | uint32(val)
	case 0x040000B6:
		d.DAD[0] = (d.DAD[0] & 0x0000FFFF) | (uint32(val) << 16)
	case 0x040000B8:
		d.WriteCNT_L(0, val)
	case 0x040000BA:
		d.WriteCNT_H(0, val)
	case 0x040000BC:
		d.SAD[1] = (d.SAD[1] & 0xFFFF0000) | uint32(val)
	case 0x040000BE:
		d.SAD[1] = (d.SAD[1] & 0x0000FFFF) | (uint32(val) << 16)
	case 0x040000C0:
		d.DAD[1] = (d.DAD[1] & 0xFFFF0000) | uint32(val)
	case 0x040000C2:
		d.DAD[1] = (d.DAD[1] & 0x0000FFFF) | (uint32(val) << 16)
	case 0x040000C4:
		d.WriteCNT_L(1, val)
	case 0x040000C6:
		d.WriteCNT_H(1, val)
	case 0x040000C8:
		d.SAD[2] = (d.SAD[2] & 0xFFFF0000) | uint32(val)
	case 0x040000CA:
		d.SAD[2] = (d.SAD[2] & 0x0000FFFF) | (uint32(val) << 16)
	case 0x040000CC:
		d.DAD[2] = (d.DAD[2] & 0xFFFF0000) | uint32(val)
	case 0x040000CE:
		d.DAD[2] = (d.DAD[2] & 0x0000FFFF) | (uint32(val) << 16)
	case 0x040000D0:
		d.WriteCNT_L(2, val)
	case 0x040000D2:
		d.WriteCNT_H(2, val)
	case 0x040000D4:
		d.SAD[3] = (d.SAD[3] & 0xFFFF0000) | uint32(val)
	case 0x040000D6:
		d.SAD[3] = (d.SAD[3] & 0x0000FFFF) | (uint32(val) << 16)
	case 0x040000D8:
		d.DAD[3] = (d.DAD[3] & 0xFFFF0000) | uint32(val)
	case 0x040000DA:
		d.DAD[3] = (d.DAD[3] & 0x0000FFFF) | (uint32(val) << 16)
	case 0x040000DC:
		d.WriteCNT_L(3, val)
	case 0x040000DE:
		d.WriteCNT_H(3, val)
	}
}

func (d *DMA) WriteRegister32(addr uint32, val uint32) {
	switch addr {
	case 0x040000B0:
		d.WriteSAD(0, val)
	case 0x040000B4:
		d.WriteDAD(0, val)
	case 0x040000BC:
		d.WriteSAD(1, val)
	case 0x040000C0:
		d.WriteDAD(1, val)
	case 0x040000C8:
		d.WriteSAD(2, val)
	case 0x040000CC:
		d.WriteDAD(2, val)
	case 0x040000D4:
		d.WriteSAD(3, val)
	case 0x040000D8:
		d.WriteDAD(3, val)
	}
}

func (d *DMA) IsActive(channel int) bool {
	if channel < 4 {
		return d.Active[channel]
	}
	return false
}
