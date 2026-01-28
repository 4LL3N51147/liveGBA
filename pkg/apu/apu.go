package apu

const (
	SampleRate = 32768
	BufferSize = 4096
)

type APU struct {
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

	WAVE_RAM [16]byte
	FIFO_A   [32]byte
	FIFO_B   [32]byte

	FIFOACount   int
	FIFO_B_Count int
	FIFO_A_Read  int
	FIFO_A_Write int
	FIFO_B_Read  int
	FIFO_B_Write int

	CycleCount  int
	SampleCount int

	SoundBuffer []int16
	BufferPos   int
}

func New() *APU {
	apu := &APU{
		SoundBuffer: make([]int16, BufferSize*2),
	}
	apu.Reset()
	return apu
}

func (a *APU) Reset() {
	a.SOUND1CNT_L = 0
	a.SOUND1CNT_H = 0
	a.SOUND1CNT_X = 0
	a.SOUND2CNT_L = 0
	a.SOUND2CNT_H = 0
	a.SOUND3CNT_L = 0
	a.SOUND3CNT_H = 0
	a.SOUND3CNT_X = 0
	a.SOUND4CNT_L = 0
	a.SOUND4CNT_H = 0
	a.SOUNDCNT_L = 0
	a.SOUNDCNT_H = 0
	a.SOUNDCNT_X = 0
	a.SOUNDBIAS = 0x0200

	a.FIFOACount = 0
	a.FIFO_B_Count = 0
	a.FIFO_A_Read = 0
	a.FIFO_A_Write = 0
	a.FIFO_B_Read = 0
	a.FIFO_B_Write = 0

	a.CycleCount = 0
	a.SampleCount = 0
	a.BufferPos = 0

	for i := range a.SoundBuffer {
		a.SoundBuffer[i] = 0
	}
}

func (a *APU) Step(cycles int) {
	a.CycleCount += cycles

	cyclesPerSample := 16777216 / SampleRate

	for a.CycleCount >= cyclesPerSample {
		a.CycleCount -= cyclesPerSample
		a.generateSample()
	}
}

func (a *APU) generateSample() {
	left := int16(0)
	right := int16(0)

	if a.SOUNDCNT_X&0x0080 != 0 {
		left = a.generateLeftChannel()
		right = a.generateRightChannel()
	}

	if a.BufferPos < len(a.SoundBuffer)-1 {
		a.SoundBuffer[a.BufferPos] = left
		a.SoundBuffer[a.BufferPos+1] = right
		a.BufferPos += 2
	}
}

func (a *APU) generateLeftChannel() int16 {
	return 0
}

func (a *APU) generateRightChannel() int16 {
	return 0
}

func (a *APU) WriteFIFO_A(data uint32) {
	for i := 0; i < 4; i++ {
		a.FIFO_A[a.FIFO_A_Write] = byte(data >> (i * 8))
		a.FIFO_A_Write = (a.FIFO_A_Write + 1) % 32
	}
	a.FIFOACount += 4
	if a.FIFOACount > 32 {
		a.FIFOACount = 32
	}
}

func (a *APU) WriteFIFO_B(data uint32) {
	for i := 0; i < 4; i++ {
		a.FIFO_B[a.FIFO_B_Write] = byte(data >> (i * 8))
		a.FIFO_B_Write = (a.FIFO_B_Write + 1) % 32
	}
	a.FIFO_B_Count += 4
	if a.FIFO_B_Count > 32 {
		a.FIFO_B_Count = 32
	}
}

func (a *APU) ReadFIFO_A() byte {
	if a.FIFOACount == 0 {
		return 0
	}
	data := a.FIFO_A[a.FIFO_A_Read]
	a.FIFO_A_Read = (a.FIFO_A_Read + 1) % 32
	a.FIFOACount--
	return data
}

func (a *APU) ReadFIFO_B() byte {
	if a.FIFO_B_Count == 0 {
		return 0
	}
	data := a.FIFO_B[a.FIFO_B_Read]
	a.FIFO_B_Read = (a.FIFO_B_Read + 1) % 32
	a.FIFO_B_Count--
	return data
}

func (a *APU) GetSamples() []int16 {
	return a.SoundBuffer[:a.BufferPos]
}

func (a *APU) ClearBuffer() {
	a.BufferPos = 0
}

func (a *APU) ReadRegister(addr uint32) uint16 {
	switch addr {
	case 0x04000060:
		return a.SOUND1CNT_L
	case 0x04000062:
		return a.SOUND1CNT_H
	case 0x04000064:
		return a.SOUND1CNT_X
	case 0x04000068:
		return a.SOUND2CNT_L
	case 0x0400006C:
		return a.SOUND2CNT_H
	case 0x04000070:
		return a.SOUND3CNT_L
	case 0x04000072:
		return a.SOUND3CNT_H
	case 0x04000074:
		return a.SOUND3CNT_X
	case 0x04000078:
		return a.SOUND4CNT_L
	case 0x0400007C:
		return a.SOUND4CNT_H
	case 0x04000080:
		return a.SOUNDCNT_L
	case 0x04000082:
		return a.SOUNDCNT_H
	case 0x04000084:
		return a.SOUNDCNT_X
	case 0x04000088:
		return a.SOUNDBIAS
	default:
		return 0
	}
}

func (a *APU) WriteRegister(addr uint32, val uint16) {
	switch addr {
	case 0x04000060:
		a.SOUND1CNT_L = val
	case 0x04000062:
		a.SOUND1CNT_H = val
	case 0x04000064:
		a.SOUND1CNT_X = val
	case 0x04000068:
		a.SOUND2CNT_L = val
	case 0x0400006C:
		a.SOUND2CNT_H = val
	case 0x04000070:
		a.SOUND3CNT_L = val
	case 0x04000072:
		a.SOUND3CNT_H = val
	case 0x04000074:
		a.SOUND3CNT_X = val
	case 0x04000078:
		a.SOUND4CNT_L = val
	case 0x0400007C:
		a.SOUND4CNT_H = val
	case 0x04000080:
		a.SOUNDCNT_L = val
	case 0x04000082:
		a.SOUNDCNT_H = val
	case 0x04000084:
		a.SOUNDCNT_X = val
	case 0x04000088:
		a.SOUNDBIAS = val
	}
}

func (a *APU) WriteWAVE_RAM(addr uint32, val uint8) {
	if addr >= 0x04000090 && addr < 0x040000A0 {
		a.WAVE_RAM[addr-0x04000090] = val
	}
}

func (a *APU) ReadWAVE_RAM(addr uint32) uint8 {
	if addr >= 0x04000090 && addr < 0x040000A0 {
		return a.WAVE_RAM[addr-0x04000090]
	}
	return 0
}
