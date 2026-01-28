package input

const (
	KeyA      = 0
	KeyB      = 1
	KeySelect = 2
	KeyStart  = 3
	KeyRight  = 4
	KeyLeft   = 5
	KeyUp     = 6
	KeyDown   = 7
	KeyR      = 8
	KeyL      = 9
)

type Input struct {
	KEYINPUT uint16
	KEYCNT   uint16

	Keys [10]bool
}

func New() *Input {
	input := &Input{}
	input.Reset()
	return input
}

func (i *Input) Reset() {
	i.KEYINPUT = 0x03FF
	i.KEYCNT = 0x0000
	for j := range i.Keys {
		i.Keys[j] = false
	}
}

func (i *Input) SetKey(key int, pressed bool) {
	if key >= 0 && key < 10 {
		i.Keys[key] = pressed
		i.updateKEYINPUT()
	}
}

func (i *Input) IsPressed(key int) bool {
	if key >= 0 && key < 10 {
		return i.Keys[key]
	}
	return false
}

func (i *Input) updateKEYINPUT() {
	i.KEYINPUT = 0x03FF

	for j := 0; j < 10; j++ {
		if i.Keys[j] {
			i.KEYINPUT &^= (1 << j)
		}
	}
}

func (i *Input) GetKEYINPUT() uint16 {
	return i.KEYINPUT
}

func (i *Input) SetKEYCNT(val uint16) {
	i.KEYCNT = val
}

func (i *Input) GetKEYCNT() uint16 {
	return i.KEYCNT
}

func (i *Input) CheckInterrupt() bool {
	if i.KEYCNT&0x4000 == 0 {
		return false
	}

	irqCondition := i.KEYCNT&0x8000 != 0
	keyMask := i.KEYCNT & 0x03FF

	keysPressed := ^i.KEYINPUT & 0x03FF

	if irqCondition {
		return (keysPressed & keyMask) == keyMask
	}

	return (keysPressed & keyMask) != 0
}

func (i *Input) WriteKEYCNT(val uint16) {
	i.KEYCNT = val
}

func (i *Input) ReadRegister(addr uint32) uint16 {
	switch addr {
	case 0x04000130:
		return i.KEYINPUT
	case 0x04000132:
		return i.KEYCNT
	}
	return 0
}

func (i *Input) WriteRegister(addr uint32, val uint16) {
	switch addr {
	case 0x04000132:
		i.WriteKEYCNT(val)
	}
}
