package memory

import (
	"fmt"
	"os"
)

// Here I initialize the struct for the memory.
// The game boy memory has 64 kbs of ram.

type Memory struct {
	bootROM        []byte
	data           [0x10000]uint8
	rom            []byte // All ROM data
	bootROMEnabled bool
	isBooting      bool
	currentBank    int
}

func New() *Memory {
	return &Memory{
		currentBank: 1,
	}
}

func (m *Memory) LoadROM(data []byte) {
	m.rom = data
	// Copy Bank 0
	for i := 0; i < 0x4000 && i < len(data); i++ {
		m.data[i] = data[i]
	}
	// Copy Bank 1 to 0x4000
	if len(data) >= 0x8000 {
		for i := 0; i < 0x4000; i++ {
			m.data[0x4000+i] = data[0x4000+i]
		}
	}
}

func (m *Memory) Read(addr uint16) uint8 {
	if m.isBooting && addr < 0x0100 {
		return m.bootROM[addr]
	}

	// Se leggiamo dall'area commutabile (0x4000 - 0x7FFF)
	if addr >= 0x4000 && addr < 0x8000 {
		if m.rom != nil {
			bankAddr := uint32(m.currentBank)*0x4000 + uint32(addr-0x4000)
			if bankAddr < uint32(len(m.rom)) {
				return m.rom[bankAddr]
			}
		}
	}

	return m.data[addr]
}

func (m *Memory) Write(address uint16, value uint8) {
	// Gestione MBC1 (molto semplificata)
	if address >= 0x2000 && address < 0x4000 {
		// Cambio banco ROM
		bank := int(value & 0x1F)
		if bank == 0 {
			bank = 1
		}
		m.currentBank = bank
		return
	}

	if address < 0x8000 {
		// Altre scritture in ROM ignorate
		return
	}
	m.data[address] = value
}

/*

	The memory is the entity that is going to Load the boot rom.
	We are going to make a function that loads the BootRom file called dmg_boot.bin.

*/

func (m *Memory) LoadBootROM(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	if len(data) != 256 {
		return fmt.Errorf("boot ROM deve essere 256 byte, ricevuti %d", len(data))
	}

	m.bootROM = data
	m.bootROMEnabled = true
	return nil
}
