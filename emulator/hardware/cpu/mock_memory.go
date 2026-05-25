package cpu

type MockMemory struct {
	Memory [0x10000]uint8
}

func NewMockMemory() *MockMemory {
	return &MockMemory{}
}

func (m *MockMemory) Read(addr uint16) uint8 {
	return m.Memory[addr]
}

func (m *MockMemory) Write(addr uint16, val uint8) {
	m.Memory[addr] = val
}
