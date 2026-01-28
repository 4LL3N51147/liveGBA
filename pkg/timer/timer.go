package timer

const (
	Prescaler1    = 0
	Prescaler64   = 1
	Prescaler256  = 2
	Prescaler1024 = 3
)

type Timer struct {
	CNT_L [4]uint16
	CNT_H [4]uint16

	Counter [4]uint32
	Reload  [4]uint16

	Prescaler [4]int
	CountUp   [4]bool
	IrqEnable [4]bool
	Enable    [4]bool

	CycleCount [4]int

	RequestInterrupt func(irq uint16)
}

func New(requestIRQ func(uint16)) *Timer {
	timer := &Timer{
		RequestInterrupt: requestIRQ,
	}
	timer.Reset()
	return timer
}

func (t *Timer) Reset() {
	for i := 0; i < 4; i++ {
		t.CNT_L[i] = 0
		t.CNT_H[i] = 0
		t.Counter[i] = 0
		t.Reload[i] = 0
		t.Prescaler[i] = 1
		t.CountUp[i] = false
		t.IrqEnable[i] = false
		t.Enable[i] = false
		t.CycleCount[i] = 0
	}
}

func (t *Timer) Step(cycles int) {
	for i := 0; i < 4; i++ {
		if !t.Enable[i] {
			continue
		}

		if t.CountUp[i] && i > 0 {
			continue
		}

		t.CycleCount[i] += cycles

		prescalerCycles := t.Prescaler[i]

		for t.CycleCount[i] >= prescalerCycles {
			t.CycleCount[i] -= prescalerCycles
			t.Counter[i]++

			if t.Counter[i] > 0xFFFF {
				t.Counter[i] = uint32(t.Reload[i])

				if t.IrqEnable[i] {
					irq := uint16(1 << (3 + i))
					t.RequestInterrupt(irq)
				}

				if i < 3 && t.CountUp[i+1] {
					t.Counter[i+1]++
					if t.Counter[i+1] > 0xFFFF {
						t.Counter[i+1] = uint32(t.Reload[i+1])
						if t.IrqEnable[i+1] {
							irq := uint16(1 << (4 + i))
							t.RequestInterrupt(irq)
						}
					}
				}
			}
		}
	}
}

func (t *Timer) WriteCNT_L(channel int, val uint16) {
	if channel >= 0 && channel < 4 {
		t.Reload[channel] = val
	}
}

func (t *Timer) WriteCNT_H(channel int, val uint16) {
	if channel < 0 || channel >= 4 {
		return
	}

	oldEnable := t.Enable[channel]
	t.CNT_H[channel] = val

	prescalerSel := val & 0x3
	switch prescalerSel {
	case Prescaler1:
		t.Prescaler[channel] = 1
	case Prescaler64:
		t.Prescaler[channel] = 64
	case Prescaler256:
		t.Prescaler[channel] = 256
	case Prescaler1024:
		t.Prescaler[channel] = 1024
	}

	t.CountUp[channel] = val&0x04 != 0
	t.IrqEnable[channel] = val&0x40 != 0
	t.Enable[channel] = val&0x80 != 0

	if !oldEnable && t.Enable[channel] {
		t.Counter[channel] = uint32(t.Reload[channel])
		t.CycleCount[channel] = 0
	}
}

func (t *Timer) ReadCNT_L(channel int) uint16 {
	if channel >= 0 && channel < 4 {
		return uint16(t.Counter[channel])
	}
	return 0
}

func (t *Timer) ReadCNT_H(channel int) uint16 {
	if channel >= 0 && channel < 4 {
		return t.CNT_H[channel]
	}
	return 0
}

func (t *Timer) ReadRegister(addr uint32) uint16 {
	switch addr {
	case 0x04000100:
		return t.ReadCNT_L(0)
	case 0x04000102:
		return t.ReadCNT_H(0)
	case 0x04000104:
		return t.ReadCNT_L(1)
	case 0x04000106:
		return t.ReadCNT_H(1)
	case 0x04000108:
		return t.ReadCNT_L(2)
	case 0x0400010A:
		return t.ReadCNT_H(2)
	case 0x0400010C:
		return t.ReadCNT_L(3)
	case 0x0400010E:
		return t.ReadCNT_H(3)
	}
	return 0
}

func (t *Timer) WriteRegister(addr uint32, val uint16) {
	switch addr {
	case 0x04000100:
		t.WriteCNT_L(0, val)
	case 0x04000102:
		t.WriteCNT_H(0, val)
	case 0x04000104:
		t.WriteCNT_L(1, val)
	case 0x04000106:
		t.WriteCNT_H(1, val)
	case 0x04000108:
		t.WriteCNT_L(2, val)
	case 0x0400010A:
		t.WriteCNT_H(2, val)
	case 0x0400010C:
		t.WriteCNT_L(3, val)
	case 0x0400010E:
		t.WriteCNT_H(3, val)
	}
}

func (t *Timer) IsEnabled(channel int) bool {
	if channel >= 0 && channel < 4 {
		return t.Enable[channel]
	}
	return false
}

func (t *Timer) GetCounter(channel int) uint16 {
	return t.ReadCNT_L(channel)
}
