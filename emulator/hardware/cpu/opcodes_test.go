package cpu

import (
	"testing"
)

type Registers struct {
	A, B, C, D, E, H, L uint8
	F                   uint8
	PC, SP              uint16
}

type TestCase struct {
	name           string
	setup          func(c *CPU)
	expected       Registers
	expectedCycles int
}

func runTests(t *testing.T, tests []TestCase) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := NewMockMemory()
			cpu := New(mem, nil)
			
			// Setup initial state
			tt.setup(cpu)

			// Step
			cycles := cpu.Step()

			// Check Registers
			if cpu.A != tt.expected.A {
				t.Errorf("Register A: expected %02X, got %02X", tt.expected.A, cpu.A)
			}
			if cpu.B != tt.expected.B {
				t.Errorf("Register B: expected %02X, got %02X", tt.expected.B, cpu.B)
			}
			if cpu.C != tt.expected.C {
				t.Errorf("Register C: expected %02X, got %02X", tt.expected.C, cpu.C)
			}
			if cpu.D != tt.expected.D {
				t.Errorf("Register D: expected %02X, got %02X", tt.expected.D, cpu.D)
			}
			if cpu.E != tt.expected.E {
				t.Errorf("Register E: expected %02X, got %02X", tt.expected.E, cpu.E)
			}
			if cpu.H != tt.expected.H {
				t.Errorf("Register H: expected %02X, got %02X", tt.expected.H, cpu.H)
			}
			if cpu.L != tt.expected.L {
				t.Errorf("Register L: expected %02X, got %02X", tt.expected.L, cpu.L)
			}
			if cpu.F != tt.expected.F {
				t.Errorf("Register F (Flags): expected %08b, got %08b", tt.expected.F, cpu.F)
			}
			if cpu.PC != tt.expected.PC {
				t.Errorf("PC: expected %04X, got %04X", tt.expected.PC, cpu.PC)
			}
			if cpu.SP != tt.expected.SP {
				t.Errorf("SP: expected %04X, got %04X", tt.expected.SP, cpu.SP)
			}
			if cycles != tt.expectedCycles {
				t.Errorf("Cycles: expected %d, got %d", tt.expectedCycles, cycles)
			}
		})
	}
}

func Test8BitLoads(t *testing.T) {
	tests := []TestCase{
		{
			name: "LD A, B (0x78)",
			setup: func(c *CPU) {
				c.PC = 0x0100
				c.Memory.Write(0x0100, 0x78)
				c.B = 0x42
			},
			expected: Registers{A: 0x42, B: 0x42, PC: 0x0101, SP: 0x0000},
			expectedCycles: 4,
		},
		{
			name: "LD B, d8 (0x06)",
			setup: func(c *CPU) {
				c.PC = 0x0100
				c.Memory.Write(0x0100, 0x06)
				c.Memory.Write(0x0101, 0xAA)
			},
			expected: Registers{B: 0xAA, PC: 0x0102, SP: 0x0000},
			expectedCycles: 8,
		},
		{
			name: "LD (HL), A (0x77)",
			setup: func(c *CPU) {
				c.PC = 0x0100
				c.Memory.Write(0x0100, 0x77)
				c.H, c.L = 0xC0, 0x00 // HL = 0xC000
				c.A = 0xDE
			},
			expected: Registers{A: 0xDE, H: 0xC0, L: 0x00, PC: 0x0101, SP: 0x0000},
			expectedCycles: 8,
		},
	}
	runTests(t, tests)
}

func Test16BitOps(t *testing.T) {
	tests := []TestCase{
		{
			name: "LD HL, d16 (0x21)",
			setup: func(c *CPU) {
				c.PC = 0x0100
				c.Memory.Write(0x0100, 0x21)
				c.Memory.Write(0x0101, 0x34)
				c.Memory.Write(0x0102, 0x12)
			},
			expected: Registers{H: 0x12, L: 0x34, PC: 0x0103, SP: 0x0000},
			expectedCycles: 12,
		},
		{
			name: "ADD HL, BC (0x09) - Half-carry",
			setup: func(c *CPU) {
				c.PC = 0x0100
				c.Memory.Write(0x0100, 0x09)
				c.H, c.L = 0x0F, 0xFF
				c.B, c.C = 0x00, 0x01
			},
			expected: Registers{
				H: 0x10, L: 0x00, 
				B: 0x00, C: 0x01,
				F: 0x20, // N=0, H=1, C=0 (Z invariato, assumiamo 0)
				PC: 0x0101,
			},
			expectedCycles: 8,
		},
	}
	runTests(t, tests)
}

func TestCBPrefix(t *testing.T) {
	tests := []TestCase{
		{
			name: "BIT 7, H (CB 7C) - Bit is 1",
			setup: func(c *CPU) {
				c.PC = 0x0100
				c.Memory.Write(0x0100, 0xCB)
				c.Memory.Write(0x0101, 0x7C)
				c.H = 0x80
			},
			expected: Registers{
				H: 0x80, 
				F: 0x20, // Z=0, N=0, H=1, C=invariato
				PC: 0x0102,
			},
			expectedCycles: 8,
		},
		{
			name: "RL C (CB 11) - Carry out",
			setup: func(c *CPU) {
				c.PC = 0x0100
				c.Memory.Write(0x0100, 0xCB)
				c.Memory.Write(0x0101, 0x11)
				c.C = 0x80
				c.setFlags(0, 0, 0, 1) // Initial Carry = 1
			},
			expected: Registers{
				C: 0x01, 
				F: 0x10, // Z=0, N=0, H=0, C=1
				PC: 0x0102,
			},
			expectedCycles: 8,
		},
		{
			name: "RES 0, A (CB 87)",
			setup: func(c *CPU) {
				c.PC = 0x0100
				c.Memory.Write(0x0100, 0xCB)
				c.Memory.Write(0x0101, 0x87)
				c.A = 0x01
			},
			expected: Registers{
				A: 0x00, 
				F: 0x00, // Flags not affected
				PC: 0x0102,
			},
			expectedCycles: 8,
		},
		{
			name: "SET 7, (HL) (CB FE)",
			setup: func(c *CPU) {
				c.PC = 0x0100
				c.Memory.Write(0x0100, 0xCB)
				c.Memory.Write(0x0101, 0xFE)
				c.H, c.L = 0xC0, 0x00
				c.Memory.Write(0xC000, 0x00)
			},
			expected: Registers{
				H: 0xC0, L: 0x00, 
				PC: 0x0102,
			},
			expectedCycles: 16,
		},
		{
			name: "SRL B (CB 38) - Result Zero",
			setup: func(c *CPU) {
				c.PC = 0x0100
				c.Memory.Write(0x0100, 0xCB)
				c.Memory.Write(0x0101, 0x38)
				c.B = 0x01
			},
			expected: Registers{
				B: 0x00, 
				F: 0x90, // Z=1, N=0, H=0, C=1
				PC: 0x0102,
			},
			expectedCycles: 8,
		},
		{
			name: "SWAP A (CB 37)",
			setup: func(c *CPU) {
				c.PC = 0x0100
				c.Memory.Write(0x0100, 0xCB)
				c.Memory.Write(0x0101, 0x37)
				c.A = 0x12
			},
			expected: Registers{
				A: 0x21, 
				F: 0x00, // Z=0
				PC: 0x0102,
			},
			expectedCycles: 8,
		},
	}
	runTests(t, tests)
}

