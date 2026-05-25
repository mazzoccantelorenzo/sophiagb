package cpu

import (
	"fmt"
	"game/hardware/gpu"
)

/* 
📜 LEGENDA TABELLA OPCODES:
---------------------------
- d8:  Dato immediato a 8 bit (un byte, 0-255).
- d16: Dato immediato a 16 bit (due byte, 0-65535).
- r8:  Offset relativo a 8 bit con segno (-128 a +127).
- a16: Indirizzo di memoria a 16 bit (0x0000-0xFFFF).
- (HL): Contenuto della memoria all'indirizzo puntato da HL.

NUMERI SOTTO L'OPCODE:
- Il primo numero indica i BYTES totali dell'istruzione (quanto avanza il PC).
- Il secondo numero indica i CYCLES di clock (la durata dell'esecuzione).
---------------------------
*/

type MemoryBus interface {
	Read(addr uint16) uint8
	Write(addr uint16, value uint8)
}

type CPU struct {
	A, F         uint8
	B, C         uint8
	D, E         uint8
	H, L         uint8
	GPU          *gpu.GPU
	SP           uint16
	PC           uint16
	IME          bool
	EnableIME    bool
	timerCounter int
	Memory       MemoryBus
}

func New(mem MemoryBus, ppu *gpu.GPU) *CPU {
	return &CPU{
		Memory: mem,
		GPU:    ppu,
		PC:     0x0000,
	}
}

func (c *CPU) setFlags(z, n, h, carry uint8) {
	c.F = (z << 7) | (n << 6) | (h << 5) | (carry << 4)
}

func (c *CPU) getReg8(index uint8) uint8 {
	switch index {
	case 0: return c.B
	case 1: return c.C
	case 2: return c.D
	case 3: return c.E
	case 4: return c.H
	case 5: return c.L
	case 6:
		hl := (uint16(c.H) << 8) | uint16(c.L)
		return c.Memory.Read(hl)
	case 7: return c.A
	default: panic(fmt.Sprintf("Invalid register index: %d", index))
	}
}

func (c *CPU) setReg8(index uint8, val uint8) {
	switch index {
	case 0: c.B = val
	case 1: c.C = val
	case 2: c.D = val
	case 3: c.E = val
	case 4: c.H = val
	case 5: c.L = val
	case 6:
		hl := (uint16(c.H) << 8) | uint16(c.L)
		c.Memory.Write(hl, val)
	case 7: c.A = val
	default: panic(fmt.Sprintf("Invalid register index: %d", index))
	}
}

func (c *CPU) jumpRelative() {
	offset := int8(c.Memory.Read(c.PC))
	c.PC++
	c.PC = uint16(int32(c.PC) + int32(offset))
}

func (c *CPU) add8(val uint8) {
	res := uint16(c.A) + uint16(val)
	h := uint8(0)
	if (c.A&0xF)+(val&0xF) > 0xF { h = 1 }
	cy := uint8(0)
	if res > 0xFF { cy = 1 }
	c.A = uint8(res)
	z := uint8(0); if c.A == 0 { z = 1 }
	c.setFlags(z, 0, h, cy)
}

func (c *CPU) adc8(val uint8) {
	carry := (c.F >> 4) & 1
	res := uint16(c.A) + uint16(val) + uint16(carry)
	h := uint8(0)
	if (c.A&0xF)+(val&0xF)+carry > 0xF { h = 1 }
	cy := uint8(0)
	if res > 0xFF { cy = 1 }
	c.A = uint8(res)
	z := uint8(0); if c.A == 0 { z = 1 }
	c.setFlags(z, 0, h, cy)
}

func (c *CPU) sub8(val uint8) {
	res := c.A - val
	h := uint8(0)
	if (c.A & 0x0F) < (val & 0x0F) { h = 1 }
	cy := uint8(0)
	if c.A < val { cy = 1 }
	c.A = res
	z := uint8(0); if c.A == 0 { z = 1 }
	c.setFlags(z, 1, h, cy)
}

func (c *CPU) sbc8(val uint8) {
	carry := (c.F >> 4) & 1
	res := int16(c.A) - int16(val) - int16(carry)
	h := uint8(0)
	if int16(c.A&0xF)-int16(val&0xF)-int16(carry) < 0 { h = 1 }
	cy := uint8(0)
	if res < 0 { cy = 1 }
	c.A = uint8(res)
	z := uint8(0); if c.A == 0 { z = 1 }
	c.setFlags(z, 1, h, cy)
}

func (c *CPU) and8(val uint8) {
	c.A &= val
	z := uint8(0); if c.A == 0 { z = 1 }
	c.setFlags(z, 0, 1, 0)
}

func (c *CPU) xor8(val uint8) {
	c.A ^= val
	z := uint8(0); if c.A == 0 { z = 1 }
	c.setFlags(z, 0, 0, 0)
}

func (c *CPU) or8(val uint8) {
	c.A |= val
	z := uint8(0); if c.A == 0 { z = 1 }
	c.setFlags(z, 0, 0, 0)
}

func (c *CPU) cp8(val uint8) {
	h := uint8(0)
	if (c.A & 0x0F) < (val & 0x0F) { h = 1 }
	cy := uint8(0)
	if c.A < val { cy = 1 }
	z := uint8(0); if c.A == val { z = 1 }
	c.setFlags(z, 1, h, cy)
}

func (c *CPU) checkInterrupts() {
	if !c.IME { return }
	ie := c.Memory.Read(0xFFFF)
	if_reg := c.Memory.Read(0xFF0F)
	if ie == 0 || if_reg == 0 { return }
	for i := uint8(0); i < 5; i++ {
		if (ie & (1 << i)) != 0 && (if_reg & (1 << i)) != 0 {
			c.executeInterrupt(i)
			return
		}
	}
}

func (c *CPU) executeInterrupt(interrupt uint8) {
	c.IME = false
	if_reg := c.Memory.Read(0xFF0F)
	if_reg &= ^(1 << interrupt)
	c.Memory.Write(0xFF0F, if_reg)
	c.SP--
	c.Memory.Write(c.SP, uint8(c.PC>>8))
	c.SP--
	c.Memory.Write(c.SP, uint8(c.PC&0xFF))
	switch interrupt {
	case 0: c.PC = 0x0040 // V-Blank
	case 1: c.PC = 0x0048 // LCD STAT
	case 2: c.PC = 0x0050 // Timer
	case 3: c.PC = 0x0058 // Serial
	case 4: c.PC = 0x0060 // Joypad
	}
}

func (c *CPU) updateTimer(cycles int) {
	c.timerCounter += cycles
	if c.timerCounter >= 256 {
		c.timerCounter -= 256
		div := c.Memory.Read(0xFF04)
		c.Memory.Write(0xFF04, div+1)
	}
}

func (c *CPU) Step() int {
	c.checkInterrupts()
	if c.EnableIME {
		c.IME = true
		c.EnableIME = false
	}

	opcode := c.Memory.Read(c.PC)
	c.PC++
	var cycles int

	switch opcode {
	case 0x00: cycles = 4 // NOP
	
	// --- 8-bit Loads ---
	case 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47,
		0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F,
		0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57,
		0x58, 0x59, 0x5A, 0x5B, 0x5C, 0x5D, 0x5E, 0x5F,
		0x60, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67,
		0x68, 0x69, 0x6A, 0x6B, 0x6C, 0x6D, 0x6E, 0x6F,
		0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x77, 0x78,
		0x79, 0x7A, 0x7B, 0x7C, 0x7D, 0x7E, 0x7F:
		dest, src := (opcode >> 3) & 0x07, opcode & 0x07
		c.setReg8(dest, c.getReg8(src))
		if src == 6 || dest == 6 { cycles = 8 } else { cycles = 4 }

	// --- 8-bit ALU (A, r) ---
	case 0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87,
		0x88, 0x89, 0x8A, 0x8B, 0x8C, 0x8D, 0x8E, 0x8F,
		0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97,
		0x98, 0x99, 0x9A, 0x9B, 0x9C, 0x9D, 0x9E, 0x9F,
		0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7,
		0xA8, 0xA9, 0xAA, 0xAB, 0xAC, 0xAD, 0xAE, 0xAF,
		0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7,
		0xB8, 0xB9, 0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF:
		val := c.getReg8(opcode & 0x07)
		if (opcode & 0x07) == 6 { cycles = 8 } else { cycles = 4 }
		switch (opcode >> 3) & 0x07 {
		case 0: c.add8(val); case 1: c.adc8(val); case 2: c.sub8(val); case 3: c.sbc8(val)
		case 4: c.and8(val); case 5: c.xor8(val); case 6: c.or8(val); case 7: c.cp8(val)
		}

	// --- 8-bit ALU (A, d8) ---
	case 0xC6, 0xCE, 0xD6, 0xDE, 0xE6, 0xEE, 0xF6, 0xFE:
		val := c.Memory.Read(c.PC); c.PC++; cycles = 8
		switch (opcode >> 3) & 0x07 {
		case 0: c.add8(val); case 1: c.adc8(val); case 2: c.sub8(val); case 3: c.sbc8(val)
		case 4: c.and8(val); case 5: c.xor8(val); case 6: c.or8(val); case 7: c.cp8(val)
		}

	// --- 16-bit Loads (rr, d16) ---
	case 0x01, 0x11, 0x21, 0x31:
		low, high := uint16(c.Memory.Read(c.PC)), uint16(c.Memory.Read(c.PC+1)); c.PC += 2; val := (high << 8) | low
		switch opcode {
		case 0x01: c.B, c.C = uint8(high), uint8(low)
		case 0x11: c.D, c.E = uint8(high), uint8(low)
		case 0x21: c.H, c.L = uint8(high), uint8(low)
		case 0x31: c.SP = val
		}
		cycles = 12

	// --- 16-bit INC/DEC ---
	case 0x03, 0x13, 0x23, 0x33: // INC rr
		switch opcode >> 4 {
		case 0: v := (uint16(c.B)<<8)|uint16(c.C); v++; c.B, c.C = uint8(v>>8), uint8(v&0xFF)
		case 1: v := (uint16(c.D)<<8)|uint16(c.E); v++; c.D, c.E = uint8(v>>8), uint8(v&0xFF)
		case 2: v := (uint16(c.H)<<8)|uint16(c.L); v++; c.H, c.L = uint8(v>>8), uint8(v&0xFF)
		case 3: c.SP++
		}
		cycles = 8
	case 0x0B, 0x1B, 0x2B, 0x3B: // DEC rr
		switch opcode >> 4 {
		case 0: v := (uint16(c.B)<<8)|uint16(c.C); v--; c.B, c.C = uint8(v>>8), uint8(v&0xFF)
		case 1: v := (uint16(c.D)<<8)|uint16(c.E); v--; c.D, c.E = uint8(v>>8), uint8(v&0xFF)
		case 2: v := (uint16(c.H)<<8)|uint16(c.L); v--; c.H, c.L = uint8(v>>8), uint8(v&0xFF)
		case 3: c.SP--
		}
		cycles = 8

	// --- 16-bit ALU (ADD HL, rr) ---
	case 0x09, 0x19, 0x29, 0x39:
		hl := uint32((uint16(c.H) << 8) | uint16(c.L)); var rr uint32
		switch opcode >> 4 {
		case 0: rr = uint32((uint16(c.B)<<8)|uint16(c.C)); case 1: rr = uint32((uint16(c.D)<<8)|uint16(c.E))
		case 2: rr = uint32(hl); case 3: rr = uint32(c.SP)
		}
		res := hl + rr; h := uint8(0); if (hl&0xFFF)+(rr&0xFFF) > 0xFFF { h = 1 }
		cy := uint8(0); if res > 0xFFFF { cy = 1 }
		c.H, c.L = uint8((res>>8)&0xFF), uint8(res&0xFF); c.setFlags((c.F>>7)&1, 0, h, cy); cycles = 8

	// --- 8-bit INC/DEC r8 ---
	case 0x04, 0x0C, 0x14, 0x1C, 0x24, 0x2C, 0x34, 0x3C: // INC r8
		reg := (opcode >> 3) & 0x07; old := c.getReg8(reg); res := old + 1; c.setReg8(reg, res)
		z := uint8(0); if res == 0 { z = 1 }; h := uint8(0); if (old & 0x0F) == 0x0F { h = 1 }
		c.setFlags(z, 0, h, (c.F>>4)&1); if reg == 6 { cycles = 12 } else { cycles = 4 }
	case 0x05, 0x0D, 0x15, 0x1D, 0x25, 0x2D, 0x35, 0x3D: // DEC r8
		reg := (opcode >> 3) & 0x07; old := c.getReg8(reg); res := old - 1; c.setReg8(reg, res)
		z := uint8(0); if res == 0 { z = 1 }; h := uint8(0); if (old & 0x0F) == 0 { h = 1 }
		c.setFlags(z, 1, h, (c.F>>4)&1); if reg == 6 { cycles = 12 } else { cycles = 4 }

	// --- Relative Jumps (JR) ---
	case 0x18: c.jumpRelative(); cycles = 12
	case 0x20, 0x28, 0x30, 0x38:
		cond := false
		switch opcode {
		case 0x20: cond = (c.F & 0x80) == 0; case 0x28: cond = (c.F & 0x80) != 0
		case 0x30: cond = (c.F & 0x10) == 0; case 0x38: cond = (c.F & 0x10) != 0
		}
		if cond { c.jumpRelative(); cycles = 12 } else { c.PC++; cycles = 8 }

	// --- Stack Operations ---
	case 0xC5, 0xD5, 0xE5, 0xF5: // PUSH rr
		var h, l uint8
		switch opcode {
		case 0xC5: h, l = c.B, c.C; case 0xD5: h, l = c.D, c.E
		case 0xE5: h, l = c.H, c.L; case 0xF5: h, l = c.A, c.F
		}
		c.SP--; c.Memory.Write(c.SP, h); c.SP--; c.Memory.Write(c.SP, l); cycles = 16
	case 0xC1, 0xD1, 0xE1, 0xF1: // POP rr
		l, h := c.Memory.Read(c.SP), c.Memory.Read(c.SP+1); c.SP += 2
		switch opcode {
		case 0xC1: c.C, c.B = l, h; case 0xD1: c.E, c.D = l, h
		case 0xE1: c.L, c.H = l, h; case 0xF1: c.F, c.A = l & 0xF0, h
		}
		cycles = 12

	// --- Returns and Jumps ---
	case 0xC9: low, high := uint16(c.Memory.Read(c.SP)), uint16(c.Memory.Read(c.SP+1)); c.SP += 2; c.PC = (high << 8) | low; cycles = 16
	case 0xD9: low, high := uint16(c.Memory.Read(c.SP)), uint16(c.Memory.Read(c.SP+1)); c.SP += 2; c.PC = (high << 8) | low; c.IME = true; cycles = 16
	case 0xC0, 0xC8, 0xD0, 0xD8: // RET cc
		cond := false
		switch opcode {
		case 0xC0: cond = (c.F & 0x80) == 0; case 0xC8: cond = (c.F & 0x80) != 0
		case 0xD0: cond = (c.F & 0x10) == 0; case 0xD8: cond = (c.F & 0x10) != 0
		}
		if cond { low, high := uint16(c.Memory.Read(c.SP)), uint16(c.Memory.Read(c.SP+1)); c.SP += 2; c.PC = (high << 8) | low; cycles = 20 } else { cycles = 8 }
	case 0xC3: low, high := uint16(c.Memory.Read(c.PC)), uint16(c.Memory.Read(c.PC+1)); c.PC = (high << 8) | low; cycles = 16
	case 0xC2, 0xCA, 0xD2, 0xDA: // JP cc, a16
		low, high := uint16(c.Memory.Read(c.PC)), uint16(c.Memory.Read(c.PC+1))
		cond := false
		switch opcode {
		case 0xC2: cond = (c.F & 0x80) == 0; case 0xCA: cond = (c.F & 0x80) != 0
		case 0xD2: cond = (c.F & 0x10) == 0; case 0xDA: cond = (c.F & 0x10) != 0
		}
		if cond { c.PC = (high << 8) | low; cycles = 16 } else { c.PC += 2; cycles = 12 }
	case 0xCD: low, high := uint16(c.Memory.Read(c.PC)), uint16(c.Memory.Read(c.PC+1)); c.PC += 2
		c.SP--; c.Memory.Write(c.SP, uint8(c.PC>>8)); c.SP--; c.Memory.Write(c.SP, uint8(c.PC&0xFF)); c.PC = (high << 8) | low; cycles = 24
	case 0xCC, 0xD4, 0xDC: // CALL cc, a16
		low, high := uint16(c.Memory.Read(c.PC)), uint16(c.Memory.Read(c.PC+1)); c.PC += 2
		cond := false
		switch opcode {
		case 0xCC: cond = (c.F & 0x80) != 0; case 0xD4: cond = (c.F & 0x10) == 0; case 0xDC: cond = (c.F & 0x10) != 0
		}
		if cond { c.SP--; c.Memory.Write(c.SP, uint8(c.PC>>8)); c.SP--; c.Memory.Write(c.SP, uint8(c.PC&0xFF)); c.PC = (high << 8) | low; cycles = 24 } else { cycles = 12 }

	// --- Memory Access ---
	case 0x02, 0x12, 0x0A, 0x1A: // LD (rr), A or LD A, (rr)
		var addr uint16
		if (opcode & 0x10) == 0 { addr = (uint16(c.B)<<8)|uint16(c.C) } else { addr = (uint16(c.D)<<8)|uint16(c.E) }
		if (opcode & 0x08) == 0 { c.Memory.Write(addr, c.A) } else { c.A = c.Memory.Read(addr) }
		cycles = 8
	case 0x22, 0x32, 0x2A, 0x3A: // HL adjust
		addr := (uint16(c.H)<<8)|uint16(c.L)
		if (opcode & 0x08) == 0 { c.Memory.Write(addr, c.A) } else { c.A = c.Memory.Read(addr) }
		if (opcode & 0x10) == 0 { addr++ } else { addr-- }
		c.H, c.L = uint8(addr>>8), uint8(addr&0xFF); cycles = 8
	case 0xE0, 0xF0: // LDH
		offset := c.Memory.Read(c.PC); c.PC++
		if opcode == 0xE0 { c.Memory.Write(0xFF00+uint16(offset), c.A) } else { c.A = c.Memory.Read(0xFF00+uint16(offset)) }
		cycles = 12
	case 0xE2, 0xF2: // LD ($FF00+C)
		if opcode == 0xE2 { c.Memory.Write(0xFF00+uint16(c.C), c.A) } else { c.A = c.Memory.Read(0xFF00+uint16(c.C)) }
		cycles = 8
	case 0xEA, 0xFA: // LD (a16), A or LD A, (a16)
		l, h := uint16(c.Memory.Read(c.PC)), uint16(c.Memory.Read(c.PC+1)); c.PC += 2; addr := (h<<8)|l
		if opcode == 0xEA { c.Memory.Write(addr, c.A) } else { c.A = c.Memory.Read(addr) }
		cycles = 16

	// --- Special / Misc ---
	case 0x06, 0x0E, 0x16, 0x1E, 0x26, 0x2E, 0x36, 0x3E: // LD r, d8
		val := c.Memory.Read(c.PC); c.PC++; reg := (opcode >> 3) & 0x07; c.setReg8(reg, val)
		if reg == 6 { cycles = 12 } else { cycles = 8 }
	case 0x07, 0x0F, 0x17, 0x1F: // Rotates A
		val, cy := c.A, (c.F>>4)&1; var nCy uint8
		switch opcode {
		case 0x07: nCy = val >> 7; c.A = (val << 1) | nCy; case 0x0F: nCy = val & 1; c.A = (val >> 1) | (nCy << 7)
		case 0x17: nCy = val >> 7; c.A = (val << 1) | cy; case 0x1F: nCy = val & 1; c.A = (val >> 1) | (cy << 7)
		}
		c.setFlags(0, 0, 0, nCy); cycles = 4
	case 0x27: // DAA
		if (c.F&0x40) == 0 { if (c.F&0x10) != 0 || c.A > 0x99 { c.A += 0x60; c.F |= 0x10 }; if (c.F&0x20) != 0 || (c.A&0x0F) > 0x09 { c.A += 0x06 }
		} else { if (c.F&0x10) != 0 { c.A -= 0x60 }; if (c.F&0x20) != 0 { c.A -= 0x06 } }
		z := uint8(0); if c.A == 0 { z = 1 }; c.F = (c.F & 0x50) | (z << 7); cycles = 4
	case 0x2F: c.A = ^c.A; c.setFlags((c.F>>7)&1, 1, 1, (c.F>>4)&1); cycles = 4
	case 0x37: c.setFlags((c.F>>7)&1, 0, 0, 1); cycles = 4 // SCF
	case 0x3F: carry := (c.F>>4)&1; nc := uint8(0); if carry == 0 { nc = 1 }; c.setFlags((c.F>>7)&1, 0, 0, nc); cycles = 4 // CCF
	case 0xF3: c.IME = false; cycles = 4 // DI
	case 0xFB: c.EnableIME = true; cycles = 4 // EI
	case 0x76: cycles = 4 // HALT
	case 0x08: l, h := uint16(c.Memory.Read(c.PC)), uint16(c.Memory.Read(c.PC+1)); c.PC += 2; addr := (h<<8)|l
		c.Memory.Write(addr, uint8(c.SP&0xFF)); c.Memory.Write(addr+1, uint8(c.SP>>8)); cycles = 20
	case 0xF9: c.SP = (uint16(c.H)<<8)|uint16(c.L); cycles = 8
	case 0xE9: c.PC = (uint16(c.H) << 8) | uint16(c.L); cycles = 4 // JP (HL)
	case 0xE8: offset := int8(c.Memory.Read(c.PC)); c.PC++; sp := c.SP; res := uint32(sp)+uint32(int32(offset))
		h := uint8(0); if (sp&0xF)+(uint16(offset)&0xF) > 0xF { h = 1 }
		cy := uint8(0); if (sp&0xFF)+(uint16(offset)&0xFF) > 0xFF { cy = 1 }
		c.SP = uint16(res); c.setFlags(0, 0, h, cy); cycles = 16
	case 0xF8: offset := int8(c.Memory.Read(c.PC)); c.PC++; sp := c.SP; res := uint32(sp)+uint32(int32(offset))
		h := uint8(0); if (sp&0xF)+(uint16(offset)&0xF) > 0xF { h = 1 }
		cy := uint8(0); if (sp&0xFF)+(uint16(offset)&0xFF) > 0xFF { cy = 1 }
		c.H, c.L = uint8(uint16(res)>>8), uint8(uint16(res)&0xFF); c.setFlags(0, 0, h, cy); cycles = 12
	case 0xC7, 0xCF, 0xD7, 0xDF, 0xE7, 0xEF, 0xF7, 0xFF: // RST
		c.SP--; c.Memory.Write(c.SP, uint8(c.PC>>8)); c.SP--; c.Memory.Write(c.SP, uint8(c.PC&0xFF))
		c.PC = uint16(opcode & 0x38); cycles = 16

	case 0xCB:
		cbOpcode := c.Memory.Read(c.PC); c.PC++; regIndex := cbOpcode & 0x07
		if regIndex == 6 { cycles = 16; if cbOpcode >= 0x40 && cbOpcode <= 0x7F { cycles = 12 } } else { cycles = 8 }
		if cbOpcode >= 0x40 && cbOpcode <= 0x7F { // BIT b, r
			bit := (cbOpcode >> 3) & 0x07; val := c.getReg8(regIndex); z := uint8(0); if (val & (1 << bit)) == 0 { z = 1 }
			c.setFlags(z, 0, 1, (c.F>>4)&1)
		} else if cbOpcode >= 0x80 && cbOpcode <= 0xBF { // RES b, r
			bit := (cbOpcode >> 3) & 0x07; val := c.getReg8(regIndex); val &= ^(1 << bit); c.setReg8(regIndex, val)
		} else if cbOpcode >= 0xC0 && cbOpcode <= 0xFF { // SET b, r
			bit := (cbOpcode >> 3) & 0x07; val := c.getReg8(regIndex); val |= (1 << bit); c.setReg8(regIndex, val)
		} else { // 0x00-0x3F: Rotations and Shifts
			val, carry := c.getReg8(regIndex), (c.F>>4)&1; var res, nCy uint8
			switch cbOpcode >> 3 {
			case 0: nCy = val >> 7; res = (val << 1) | nCy; case 1: nCy = val & 1; res = (val >> 1) | (nCy << 7)
			case 2: nCy = val >> 7; res = (val << 1) | carry; case 3: nCy = val & 1; res = (val >> 1) | (carry << 7)
			case 4: nCy = val >> 7; res = val << 1; case 5: nCy = val & 1; res = (val >> 1) | (val & 0x80)
			case 6: res = (val << 4) | (val >> 4); nCy = 0; case 7: nCy = val & 1; res = val >> 1
			}
			z := uint8(0); if res == 0 { z = 1 }; if (cbOpcode >> 3) == 6 { c.setFlags(z, 0, 0, 0) } else { c.setFlags(z, 0, 0, nCy) }
			c.setReg8(regIndex, res)
		}

	default:
		panic(fmt.Sprintf("PANIC: Opcode 0x%02X non implementato al PC: 0x%04X. I registri erano: A:%02X BC:%02X%02X HL:%02X%02X",
			opcode, c.PC-1, c.A, c.B, c.C, c.H, c.L))
	}

	c.updateTimer(cycles)
	if c.GPU != nil { c.GPU.Step(cycles) }
	return cycles
}
