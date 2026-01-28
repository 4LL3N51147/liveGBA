package cpu

func (c *CPU) thumbAddSub(instr uint32) int {
	op := (instr >> 9) & 1
	rnOffset := (instr >> 6) & 7
	rs := (instr >> 3) & 7
	rd := instr & 7

	rsVal := c.Regs[rs]
	var operand uint32

	if (instr & 0x0400) != 0 {
		operand = uint32(rnOffset)
	} else {
		operand = c.Regs[rnOffset]
	}

	var result uint32
	if op == 0 {
		result = rsVal + operand
		c.SetFlag(FlagC, result < rsVal)
		c.SetFlag(FlagV, ((rsVal^operand)&0x80000000) == 0 && ((rsVal^result)&0x80000000) != 0)
	} else {
		result = rsVal - operand
		c.SetFlag(FlagC, rsVal >= operand)
		c.SetFlag(FlagV, ((rsVal^operand)&0x80000000) != 0 && ((rsVal^result)&0x80000000) != 0)
	}

	c.SetNZ(result)
	c.Regs[rd] = result

	return 1
}

func (c *CPU) thumbMoveShifted(instr uint32) int {
	op := (instr >> 11) & 3
	offset := (instr >> 6) & 0x1F
	rs := (instr >> 3) & 7
	rd := instr & 7

	value := c.Regs[rs]
	var result uint32
	var carry bool

	switch op {
	case 0:
		if offset == 0 {
			result = value
			carry = c.GetFlag(FlagC)
		} else {
			result = value << offset
			carry = ((value >> (32 - offset)) & 1) != 0
		}
	case 1:
		if offset == 0 {
			offset = 32
		}
		result = value >> offset
		carry = ((value >> (offset - 1)) & 1) != 0
	case 2:
		if offset == 0 {
			offset = 32
		}
		sign := value & 0x80000000
		result = value >> offset
		if sign != 0 {
			result |= ^(uint32(0xFFFFFFFF) >> offset)
		}
		carry = ((value >> (offset - 1)) & 1) != 0
	case 3:
		if offset == 0 {
			result = value
			carry = c.GetFlag(FlagC)
		} else {
			result = (value >> offset) | (value << (32 - offset))
			carry = ((value >> (offset - 1)) & 1) != 0
		}
	}

	c.SetNZC(result, carry)
	c.Regs[rd] = result

	return 1
}

func (c *CPU) thumbMoveCompare(instr uint32) int {
	op := (instr >> 11) & 3
	rd := (instr >> 8) & 7
	offset := instr & 0xFF

	value := c.Regs[rd]
	var result uint32

	switch op {
	case 0:
		result = uint32(offset)
		c.Regs[rd] = result
	case 1:
		result = value - uint32(offset)
		c.SetFlag(FlagC, value >= uint32(offset))
		c.SetFlag(FlagV, ((value^uint32(offset))&0x80000000) != 0 && ((value^result)&0x80000000) != 0)
	case 2:
		result = value + uint32(offset)
		c.SetFlag(FlagC, result < value)
		c.SetFlag(FlagV, ((value^uint32(offset))&0x80000000) == 0 && ((value^result)&0x80000000) != 0)
		c.Regs[rd] = result
	case 3:
		result = value - uint32(offset)
		c.SetFlag(FlagC, value >= uint32(offset))
		c.SetFlag(FlagV, ((value^uint32(offset))&0x80000000) != 0 && ((value^result)&0x80000000) != 0)
	}

	c.SetNZ(result)

	return 1
}

func (c *CPU) thumbALU(instr uint32) int {
	op := (instr >> 6) & 0xF
	rs := (instr >> 3) & 7
	rd := instr & 7

	left := c.Regs[rd]
	right := c.Regs[rs]
	var result uint32
	var carry bool
	var overflow bool

	switch op {
	case 0x0:
		result = left & right
	case 0x1:
		result = left ^ right
	case 0x2:
		shift := right & 0xFF
		if shift == 0 {
			result = left
			carry = c.GetFlag(FlagC)
		} else if shift < 32 {
			result = left << shift
			carry = ((left >> (32 - shift)) & 1) != 0
		} else {
			result = 0
			carry = false
		}
	case 0x3:
		shift := right & 0xFF
		if shift == 0 {
			result = left
			carry = c.GetFlag(FlagC)
		} else if shift < 32 {
			result = left >> shift
			carry = ((left >> (shift - 1)) & 1) != 0
		} else {
			result = 0
			carry = (left & 0x80000000) != 0
		}
	case 0x4:
		shift := right & 0xFF
		if shift == 0 {
			result = left
			carry = c.GetFlag(FlagC)
		} else if shift >= 32 {
			if left&0x80000000 != 0 {
				result = 0xFFFFFFFF
				carry = true
			} else {
				result = 0
				carry = false
			}
		} else {
			sign := left & 0x80000000
			result = left >> shift
			if sign != 0 {
				result |= ^(uint32(0xFFFFFFFF) >> shift)
			}
			carry = ((left >> (shift - 1)) & 1) != 0
		}
	case 0x5:
		shift := right & 0x1F
		if shift == 0 {
			result = left
			carry = c.GetFlag(FlagC)
		} else {
			result = (left >> shift) | (left << (32 - shift))
			carry = ((left >> (shift - 1)) & 1) != 0
		}
	case 0x6:
		result = left & right
		c.SetFlag(FlagC, (left&(1<<(right&0x1F))) != 0)
	case 0x7:
		result = left &^ right
	case 0x8:
		result = left - right
		carry = left >= right
		overflow = ((left^right)&0x80000000) != 0 && ((left^result)&0x80000000) != 0
	case 0x9:
		result = left + right
		carry = result < left
		overflow = ((left^right)&0x80000000) == 0 && ((left^result)&0x80000000) != 0
	case 0xA:
		result = left - right
		carry = left >= right
		overflow = ((left^right)&0x80000000) != 0 && ((left^result)&0x80000000) != 0
	case 0xB:
		result = left + right
		carry = result < left
		overflow = ((left^right)&0x80000000) == 0 && ((left^result)&0x80000000) != 0
	case 0xC:
		result = left | right
	case 0xD:
		result = right
	case 0xE:
		result = left &^ right
	case 0xF:
		result = ^right
	}

	c.SetNZCV(result, carry, overflow)
	if op != 0x8 && op != 0xA && op != 0xB {
		c.Regs[rd] = result
	}

	return 1
}

func (c *CPU) thumbHiReg(instr uint32) int {
	op := (instr >> 8) & 3
	h1 := (instr >> 7) & 1
	h2 := (instr >> 6) & 1
	rs := (instr >> 3) & 7
	rd := instr & 7

	if h2 != 0 {
		rs += 8
	}
	if h1 != 0 {
		rd += 8
	}

	rsVal := c.GetReg(int(rs))
	rdVal := c.GetReg(int(rd))

	switch op {
	case 0x0:
		result := rdVal + rsVal
		c.SetReg(int(rd), result)
		if rd == 15 {
			c.PC = result &^ 1
			c.Regs[15] = c.PC + 2
		}
	case 0x1:
		result := rdVal - rsVal
		c.SetFlag(FlagN, result&0x80000000 != 0)
		c.SetFlag(FlagZ, result == 0)
		c.SetFlag(FlagC, rdVal >= rsVal)
		c.SetFlag(FlagV, ((rdVal^rsVal)&0x80000000) != 0 && ((rdVal^result)&0x80000032) != 0)
	case 0x2:
		c.SetReg(int(rd), rsVal)
		if rd == 15 {
			c.PC = rsVal &^ 1
			c.Regs[15] = c.PC + 2
		}
	case 0x3:
		c.PC = rsVal &^ 1
		c.Regs[15] = c.PC + 2
		c.CPSR |= FlagT
	}

	return 1
}

func (c *CPU) thumbLoadPC(instr uint32) int {
	rd := (instr >> 8) & 7
	offset := (instr & 0xFF) << 2

	addr := (c.PC &^ 3) + uint32(offset)
	c.Regs[rd] = c.Read32(addr)

	return 1
}

func (c *CPU) thumbLoadStoreReg(instr uint32) int {
	return 1
}

func (c *CPU) thumbLoadStoreSign(instr uint32) int {
	return 1
}

func (c *CPU) thumbLoadStoreImm(instr uint32) int {
	return 1
}

func (c *CPU) thumbLoadStoreH(instr uint32) int {
	return 1
}

func (c *CPU) thumbLoadStoreSP(instr uint32) int {
	return 1
}

func (c *CPU) thumbLoadAddr(instr uint32) int {
	return 1
}

func (c *CPU) thumbAddSP(instr uint32) int {
	return 1
}

func (c *CPU) thumbPushPop(instr uint32) int {
	return 1
}

func (c *CPU) thumbBlockTransfer(instr uint32) int {
	return 1
}

func (c *CPU) thumbSWI(instr uint32) int {
	return 1
}

func (c *CPU) thumbCondBranch(instr uint32) int {
	return 1
}

func (c *CPU) thumbUncondBranch(instr uint32) int {
	return 1
}

func (c *CPU) thumbLongBranch(instr uint32) int {
	return 1
}
