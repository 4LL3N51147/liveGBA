package cpu

import (
	"fmt"
)

const (
	ModeUser       = 0x10
	ModeFIQ        = 0x11
	ModeIRQ        = 0x12
	ModeSupervisor = 0x13
	ModeAbort      = 0x17
	ModeUndefined  = 0x1B
	ModeSystem     = 0x1F
)

const (
	FlagN = 1 << 31 // Negative
	FlagZ = 1 << 30 // Zero
	FlagC = 1 << 29 // Carry
	FlagV = 1 << 28 // Overflow
	FlagI = 1 << 7  // IRQ disable
	FlagF = 1 << 6  // FIQ disable
	FlagT = 1 << 5  // Thumb mode
)

type CPU struct {
	Regs    [16]uint32
	RegsFIQ [7]uint32 // R8-R14 FIQ mode
	RegsIRQ [2]uint32 // R13-R14 IRQ mode
	RegsSVC [2]uint32 // R13-R14 Supervisor mode
	RegsABT [2]uint32 // R13-R14 Abort mode
	RegsUND [2]uint32 // R13-R14 Undefined mode
	CPSR    uint32
	SPSR    uint32
	SPSRfiq uint32
	SPSRirq uint32
	SPSRsvc uint32
	SPSRabt uint32
	SPSRund uint32

	PC uint32
	LR uint32
	SP uint32

	Pipeline  [3]uint32
	Halted    bool
	StepCount uint64

	Read8   func(addr uint32) uint8
	Read16  func(addr uint32) uint16
	Read32  func(addr uint32) uint32
	Write8  func(addr uint32, val uint8)
	Write16 func(addr uint32, val uint16)
	Write32 func(addr uint32, val uint32)
}

func New() *CPU {
	cpu := &CPU{}
	cpu.Reset()
	return cpu
}

func (c *CPU) Reset() {
	for i := range c.Regs {
		c.Regs[i] = 0
	}
	// 初始化栈指针（模拟 BIOS 的行为）
	c.Regs[13] = 0x03007F00   // 用户模式 SP
	c.RegsSVC[0] = 0x03007FE0 // SVC 模式 SP
	c.RegsIRQ[0] = 0x03007FA0 // IRQ 模式 SP

	c.CPSR = ModeSupervisor | FlagI | FlagF
	c.UpdateMode()
	c.PC = 0
	c.Pipeline[0] = 0
	c.Pipeline[1] = 0
	c.Pipeline[2] = 0
	c.Halted = false
	c.StepCount = 0
}

func (c *CPU) UpdateMode() {
	mode := c.CPSR & 0x1F
	switch mode {
	case ModeFIQ:
		c.SP = c.RegsFIQ[5]
		c.LR = c.RegsFIQ[6]
	case ModeIRQ:
		c.SP = c.RegsIRQ[0]
		c.LR = c.RegsIRQ[1]
	case ModeSupervisor:
		c.SP = c.RegsSVC[0]
		c.LR = c.RegsSVC[1]
	case ModeAbort:
		c.SP = c.RegsABT[0]
		c.LR = c.RegsABT[1]
	case ModeUndefined:
		c.SP = c.RegsUND[0]
		c.LR = c.RegsUND[1]
	default:
		c.SP = c.Regs[13]
		c.LR = c.Regs[14]
	}
	c.Regs[15] = c.PC
}

func (c *CPU) SaveMode() {
	mode := c.CPSR & 0x1F
	switch mode {
	case ModeFIQ:
		c.RegsFIQ[5] = c.SP
		c.RegsFIQ[6] = c.LR
	case ModeIRQ:
		c.RegsIRQ[0] = c.SP
		c.RegsIRQ[1] = c.LR
	case ModeSupervisor:
		c.RegsSVC[0] = c.SP
		c.RegsSVC[1] = c.LR
	case ModeAbort:
		c.RegsABT[0] = c.SP
		c.RegsABT[1] = c.LR
	case ModeUndefined:
		c.RegsUND[0] = c.SP
		c.RegsUND[1] = c.LR
	default:
		c.Regs[13] = c.SP
		c.Regs[14] = c.LR
	}
}

func (c *CPU) SwitchMode(newMode uint32) {
	c.SaveMode()
	oldMode := c.CPSR & 0x1F
	c.CPSR = (c.CPSR &^ 0x1F) | newMode

	switch oldMode {
	case ModeFIQ:
		c.SPSRfiq = c.CPSR
	case ModeIRQ:
		c.SPSRirq = c.CPSR
	case ModeSupervisor:
		c.SPSRsvc = c.CPSR
	case ModeAbort:
		c.SPSRabt = c.CPSR
	case ModeUndefined:
		c.SPSRund = c.CPSR
	}

	c.UpdateMode()
}

func (c *CPU) GetSPSR() uint32 {
	switch c.CPSR & 0x1F {
	case ModeFIQ:
		return c.SPSRfiq
	case ModeIRQ:
		return c.SPSRirq
	case ModeSupervisor:
		return c.SPSRsvc
	case ModeAbort:
		return c.SPSRabt
	case ModeUndefined:
		return c.SPSRund
	default:
		return c.CPSR
	}
}

func (c *CPU) SetSPSR(val uint32) {
	switch c.CPSR & 0x1F {
	case ModeFIQ:
		c.SPSRfiq = val
	case ModeIRQ:
		c.SPSRirq = val
	case ModeSupervisor:
		c.SPSRsvc = val
	case ModeAbort:
		c.SPSRabt = val
	case ModeUndefined:
		c.SPSRund = val
	}
}

func (c *CPU) GetReg(n int) uint32 {
	if n == 15 {
		return c.PC
	}
	return c.Regs[n]
}

func (c *CPU) SetReg(n int, val uint32) {
	if n == 15 {
		c.PC = val &^ 1
	} else {
		c.Regs[n] = val
	}
}

func (c *CPU) InThumbMode() bool {
	return c.CPSR&FlagT != 0
}

func (c *CPU) SetFlag(flag uint32, set bool) {
	if set {
		c.CPSR |= flag
	} else {
		c.CPSR &= ^flag
	}
}

func (c *CPU) GetFlag(flag uint32) bool {
	return c.CPSR&flag != 0
}

func (c *CPU) SetNZ(val uint32) {
	c.SetFlag(FlagN, val&0x80000000 != 0)
	c.SetFlag(FlagZ, val == 0)
}

func (c *CPU) SetNZC(val uint32, carry bool) {
	c.SetNZ(val)
	c.SetFlag(FlagC, carry)
}

func (c *CPU) SetNZCV(val uint32, carry, overflow bool) {
	c.SetNZC(val, carry)
	c.SetFlag(FlagV, overflow)
}

func (c *CPU) ConditionPassed(cond uint32) bool {
	n := c.GetFlag(FlagN)
	z := c.GetFlag(FlagZ)
	cv := c.GetFlag(FlagC)
	v := c.GetFlag(FlagV)

	switch cond {
	case 0x0:
		return z
	case 0x1:
		return !z
	case 0x2:
		return cv
	case 0x3:
		return !cv
	case 0x4:
		return n
	case 0x5:
		return !n
	case 0x6:
		return v
	case 0x7:
		return !v
	case 0x8:
		return cv && !z
	case 0x9:
		return !cv || z
	case 0xA:
		return n == v
	case 0xB:
		return n != v
	case 0xC:
		return !z && (n == v)
	case 0xD:
		return z || (n != v)
	case 0xE:
		return true
	default:
		return false
	}
}

func (c *CPU) Step() int {
	if c.Halted {
		return 1
	}

	oldPC := c.PC
	cycles := 1

	if c.InThumbMode() {
		cycles = c.executeThumb()
	} else {
		cycles = c.executeARM()
	}

	// 检查 PC 是否异常变化
	if oldPC >= 0x08000000 && c.PC < 0x01000000 && c.StepCount < 100 {
		fmt.Printf("[CPU-ERROR] PC jumped from 0x%08X to 0x%08X at step %d!\n",
			oldPC, c.PC, c.StepCount)
	}

	c.StepCount++
	// 每 1000000 步输出一次 PC
	if c.StepCount%1000000 == 0 {
		fmt.Printf("[CPU] Step %d, PC: 0x%08X, Mode: %s\n",
			c.StepCount, c.PC, c.getModeString())
	}

	return cycles
}

func (c *CPU) executeARM() int {
	instr := c.Read32(c.PC)

	// 每 100000 步输出详细执行信息
	if c.StepCount%100000 == 0 {
		fmt.Printf("[CPU-DEBUG] PC: 0x%08X, Instr: 0x%08X\n", c.PC, instr)
	}

	// 先检查条件
	cond := (instr >> 28) & 0xF
	if !c.ConditionPassed(cond) {
		c.PC += 4
		c.Regs[15] = c.PC + 4
		return 1
	}

	switch {
	case (instr&0x0E000000) == 0x00000000 && (instr&0x00000010) != 0x00000010:
		if c.PC < 0x100 {
			fmt.Printf("[BIOS] DataProcessing at PC=0x%08X, instr=0x%08X, R9=0x%08X\n", c.PC, instr, c.Regs[9])
		}
		c.PC += 4
		c.Regs[15] = c.PC + 4
		return c.handleDataProcessing(instr)
	case (instr & 0x0E000000) == 0x02000000:
		if c.StepCount < 5 {
			fmt.Printf("[CPU] PSRTransfer at PC=0x%08X\n", c.PC)
		}
		c.PC += 4
		c.Regs[15] = c.PC + 4
		return c.handlePSRTransfer(instr)
	case (instr & 0x0E000090) == 0x00000090:
		if c.StepCount < 5 {
			fmt.Printf("[CPU] Multiply at PC=0x%08X\n", c.PC)
		}
		c.PC += 4
		c.Regs[15] = c.PC + 4
		return c.handleMultiply(instr)
	case (instr&0x0E000F00) == 0x000000B0 || (instr&0x0E000F00) == 0x000000D0:
		if c.StepCount < 5 {
			fmt.Printf("[CPU] HalfwordTransfer at PC=0x%08X\n", c.PC)
		}
		c.PC += 4
		c.Regs[15] = c.PC + 4
		return c.handleHalfwordTransfer(instr)
	case (instr & 0x0E000000) == 0x08000000:
		// Block Transfer (LDM/STM): bit 27:25 = 100
		if c.PC < 0x100 {
			fmt.Printf("[BIOS] BlockTransfer at PC=0x%08X, instr=0x%08X, SP=0x%08X, R9=0x%08X\n", c.PC, instr, c.Regs[13], c.Regs[9])
		}
		c.PC += 4
		c.Regs[15] = c.PC + 4
		return c.handleBlockTransfer(instr)
	case (instr & 0x0E000000) == 0x04000000:
		if c.PC < 0x100 {
			fmt.Printf("[BIOS] SingleTransfer at PC=0x%08X, instr=0x%08X, Rn=R%d\n", c.PC, instr, (instr>>16)&0xF)
		}
		c.PC += 4
		c.Regs[15] = c.PC + 4
		return c.handleSingleTransfer(instr)
	case (instr & 0x0E000000) == 0x0A000000:
		// Branch 指令: 位 27:25 = 101
		// 使用当前 PC（不是 PC+4）计算目标
		if c.PC < 0x100 {
			fmt.Printf("[BIOS] Branch at PC=0x%08X, instr=0x%08X\n", c.PC, instr)
		}
		cycles := c.handleBranch(instr)
		if c.PC >= 0x08000000 {
			fmt.Printf("[BIOS] Jumped to ROM! PC=0x%08X\n", c.PC)
		}
		return cycles
	case (instr & 0x0F000000) == 0x0F000000:
		c.PC += 4
		c.Regs[15] = c.PC + 4
		return c.handleSWI(instr)
	case (instr & 0x0E000000) == 0x0C000000:
		c.PC += 4
		c.Regs[15] = c.PC + 4
		return c.handleCoprocessorTransfer(instr)
	default:
		if c.StepCount < 10 {
			fmt.Printf("[CPU-DEFAULT] PC=0x%08X, Instr=0x%08X, opcode=0x%02X\n",
				c.PC, instr, (instr>>25)&0x7)
		}
		c.PC += 4
		c.Regs[15] = c.PC + 4
		return 1
	}
}

func (c *CPU) executeThumb() int {
	instr := uint32(c.Read16(c.PC))
	c.PC += 2
	c.Regs[15] = c.PC + 2

	switch {
	case (instr & 0xF800) == 0x1800:
		return c.thumbAddSub(instr)
	case (instr & 0xE000) == 0x0000:
		return c.thumbMoveShifted(instr)
	case (instr & 0xE000) == 0x2000:
		return c.thumbMoveCompare(instr)
	case (instr & 0xFC00) == 0x4000:
		return c.thumbALU(instr)
	case (instr & 0xFC00) == 0x4400:
		return c.thumbHiReg(instr)
	case (instr & 0xF800) == 0x4800:
		return c.thumbLoadPC(instr)
	case (instr & 0xF200) == 0x5000:
		return c.thumbLoadStoreReg(instr)
	case (instr & 0xF200) == 0x5200:
		return c.thumbLoadStoreSign(instr)
	case (instr & 0xE000) == 0x6000:
		return c.thumbLoadStoreImm(instr)
	case (instr & 0xF000) == 0x8000:
		return c.thumbLoadStoreH(instr)
	case (instr & 0xF000) == 0x9000:
		return c.thumbLoadStoreSP(instr)
	case (instr & 0xF000) == 0xA000:
		return c.thumbLoadAddr(instr)
	case (instr & 0xFF00) == 0xB000:
		return c.thumbAddSP(instr)
	case (instr & 0xF600) == 0xB400:
		return c.thumbPushPop(instr)
	case (instr & 0xF000) == 0xC000:
		return c.thumbBlockTransfer(instr)
	case (instr & 0xFF00) == 0xDF00:
		return c.thumbSWI(instr)
	case (instr & 0xF000) == 0xD000:
		return c.thumbCondBranch(instr)
	case (instr & 0xF800) == 0xE000:
		return c.thumbUncondBranch(instr)
	case (instr & 0xF800) == 0xE800:
		return c.thumbLongBranch(instr)
	default:
		return 1
	}
}

func (c *CPU) handleDataProcessing(instr uint32) int {
	opcode := (instr >> 21) & 0xF
	s := (instr >> 20) & 1
	rn := (instr >> 16) & 0xF
	rd := (instr >> 12) & 0xF

	var operand2 uint32
	var carry bool

	if (instr & 0x02000000) != 0 {
		imm := instr & 0xFF
		rot := ((instr >> 8) & 0xF) * 2
		operand2 = (imm >> rot) | (imm << (32 - rot))
		if rot != 0 {
			carry = (operand2 & 0x80000000) != 0
		} else {
			carry = c.GetFlag(FlagC)
		}
	} else {
		rm := instr & 0xF
		operand2 = c.Regs[rm]
		shiftType := (instr >> 5) & 3

		if (instr & 0x10) != 0 {
			shift := c.Regs[(instr>>8)&0xF] & 0xFF
			operand2, carry = c.shift(shiftType, operand2, shift)
		} else {
			shift := (instr >> 7) & 0x1F
			if shift == 0 {
				shift = 32
			}
			operand2, carry = c.shift(shiftType, operand2, shift)
		}
	}

	rnVal := c.Regs[rn]
	if rn == 15 {
		rnVal += 4
	}

	// 调试：检查 R9 的值
	if rn == 9 && c.StepCount < 10 {
		fmt.Printf("[CPU-DEBUG] DataProcessing: rn=R%d, rnVal=0x%08X, Rd=R%d\n", rn, rnVal, rd)
	}

	var result uint32
	var resultCarry bool
	var overflow bool

	switch opcode {
	case 0x0:
		result = rnVal & operand2
	case 0x1:
		result = rnVal ^ operand2
	case 0x2:
		result = rnVal - operand2
		resultCarry = rnVal >= operand2
		overflow = ((rnVal^operand2)&0x80000000) != 0 && ((rnVal^result)&0x80000000) != 0
	case 0x3:
		result = operand2 - rnVal
		resultCarry = operand2 >= rnVal
		overflow = ((operand2^rnVal)&0x80000000) != 0 && ((operand2^result)&0x80000000) != 0
	case 0x4:
		result = rnVal + operand2
		resultCarry = result < rnVal
		overflow = ((rnVal^operand2)&0x80000000) == 0 && ((rnVal^result)&0x80000000) != 0
	case 0x5:
		result = rnVal + operand2 + boolToUint32(c.GetFlag(FlagC))
		resultCarry = result < rnVal || (result == rnVal && c.GetFlag(FlagC))
	case 0x6:
		result = rnVal - operand2 - boolToUint32(!c.GetFlag(FlagC))
		resultCarry = rnVal > operand2 || (rnVal == operand2 && c.GetFlag(FlagC))
	case 0x7:
		result = operand2 - rnVal - boolToUint32(!c.GetFlag(FlagC))
		resultCarry = operand2 > rnVal || (operand2 == rnVal && c.GetFlag(FlagC))
	case 0x8:
		result = rnVal & operand2
		c.SetFlag(FlagC, carry)
	case 0x9:
		result = rnVal ^ operand2
		c.SetFlag(FlagC, carry)
	case 0xA:
		result = rnVal - operand2
		resultCarry = rnVal >= operand2
		overflow = ((rnVal^operand2)&0x80000000) != 0 && ((rnVal^result)&0x80000000) != 0
	case 0xB:
		result = rnVal + operand2
		resultCarry = result < rnVal
		overflow = ((rnVal^operand2)&0x80000000) == 0 && ((rnVal^result)&0x80000000) != 0
	case 0xC:
		result = rnVal | operand2
	case 0xD:
		result = operand2
	case 0xE:
		result = rnVal &^ operand2
	case 0xF:
		result = ^operand2
	}

	if s == 1 && rd != 15 {
		if opcode >= 0x8 && opcode <= 0xB {
			c.SetNZCV(result, resultCarry, overflow)
		} else if opcode >= 0x2 && opcode <= 0x7 {
			c.SetNZCV(result, resultCarry, overflow)
		} else {
			c.SetNZ(result)
		}
	}

	if opcode != 0x8 && opcode != 0x9 && opcode != 0xA && opcode != 0xB {
		c.Regs[rd] = result
	}

	if rd == 15 {
		if s == 1 {
			c.CPSR = c.GetSPSR()
			c.UpdateMode()
		}
		c.PC = result &^ 3
		c.Regs[15] = c.PC + 4
	}

	return 1
}

func (c *CPU) handlePSRTransfer(instr uint32) int {
	return 1
}

func (c *CPU) handleMultiply(instr uint32) int {
	return 1
}

func (c *CPU) handleHalfwordTransfer(instr uint32) int {
	return 1
}

func (c *CPU) handleSingleTransfer(instr uint32) int {
	return 1
}

func (c *CPU) handleBlockTransfer(instr uint32) int {
	return 1
}

func (c *CPU) handleBranch(instr uint32) int {
	// 分支指令: B/BL
	// 位 27:25 = 101
	// 位 24 = L (Link)
	// 位 23:0 = 偏移量 (24位有符号数)

	link := (instr >> 24) & 1
	offset := instr & 0x00FFFFFF

	// 符号扩展 24 位偏移量到 32 位
	if offset&0x00800000 != 0 {
		offset |= 0xFF000000
	}

	// 偏移量左移 2 位 (ARM 指令 4 字节对齐)
	offset <<= 2

	// 保存返回地址 (如果需要)
	if link == 1 {
		c.LR = c.PC - 4
	}

	// 计算目标地址: PC + 8 (流水线) + 偏移量
	target := c.PC + offset
	c.PC = target
	c.Regs[15] = c.PC + 4

	return 1
}

func (c *CPU) handleSWI(instr uint32) int {
	return 1
}

func (c *CPU) handleCoprocessorTransfer(instr uint32) int {
	return 1
}

func (c *CPU) shift(shiftType uint32, value uint32, shift uint32) (uint32, bool) {
	if shift == 0 {
		return value, c.GetFlag(FlagC)
	}

	switch shiftType {
	case 0:
		if shift >= 32 {
			return 0, (value & 1) != 0
		}
		return value << shift, ((value >> (32 - shift)) & 1) != 0
	case 1:
		if shift >= 32 {
			return 0, (value & 0x80000000) != 0
		}
		return value >> shift, ((value >> (shift - 1)) & 1) != 0
	case 2:
		if shift >= 32 {
			if value&0x80000000 != 0 {
				return 0xFFFFFFFF, true
			}
			return 0, false
		}
		sign := value & 0x80000000
		result := value >> shift
		if sign != 0 {
			result |= ^(uint32(0xFFFFFFFF) >> shift)
		}
		return result, ((value >> (shift - 1)) & 1) != 0
	case 3:
		shift = shift & 31
		if shift == 0 {
			return value, c.GetFlag(FlagC)
		}
		return (value >> shift) | (value << (32 - shift)), ((value >> (shift - 1)) & 1) != 0
	}
	return value, false
}

func boolToUint32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func (c *CPU) getModeString() string {
	mode := c.CPSR & 0x1F
	switch mode {
	case ModeUser:
		return "USR"
	case ModeFIQ:
		return "FIQ"
	case ModeIRQ:
		return "IRQ"
	case ModeSupervisor:
		return "SVC"
	case ModeAbort:
		return "ABT"
	case ModeUndefined:
		return "UND"
	case ModeSystem:
		return "SYS"
	default:
		return fmt.Sprintf("UNK(%02X)", mode)
	}
}

func (c *CPU) String() string {
	return fmt.Sprintf("PC: %08X CPSR: %08X", c.PC, c.CPSR)
}
