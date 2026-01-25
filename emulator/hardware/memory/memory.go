package memory

import (
	"fmt"
	"os"
)

// Here I initialize the struct for the memory.
// The game boy memory has 64 kbs of ram.

type Memory struct {
	bootROM []byte         // This is memory for the rom
	data    [0x10000]uint8 //We have 64 kbs , from 0x0000 to 0xFFFF.
}

// I create the memory through this function. I can inject it in Main
func New() *Memory {
	return &Memory{}
}

func (m *Memory) Read(address uint16) uint8 {
	return m.data[address]
}

func (m *Memory) Write(address uint16, value uint8) {
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
	// m.bootROMEnabled = true
	return nil
}
