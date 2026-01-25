package cpu

// Here we are going to organize all the information related on the cpu
// The bible to check for documentation is https://gbdev.io/pandocs/CPU_Registers_and_Flags.html
/* The gameboy cpu is composed by 8 registers at 8 bits.

 	A -> Every math operation is done here and this is considered as a default register

	In the docs infact, you can see that low code instructions can omit the register A as it is considered default:
	All arithmetic and logic instructions that use register A as a destination can omit the destination, since it is assumed to be register A by default. So the following two lines have the same effect:
		"OR A,B"
		"OR B"	 (Same as A,B).

  B , C , D , E , H , L -> 	Generic use, some instructions use these in a specific way

  F  ->  Special register, the cpu updates this after every operation.
  SP ->  Stack pointer,
  PC ->  It has the next instruction's address
*/

// MemoryBus is the interface that i use so i can inject the memory.
type MemoryBus interface {
	Read(addr uint16) uint8
	Write(addr uint16, value uint8)
}

// We can think about the cpu through a struct.
type CPU struct {
	// 8 bit registers
	A, F uint8
	B, C uint8
	D, E uint8
	H, L uint8

	// 16-bit registers
	SP     uint16    // Stack Pointer
	PC     uint16    // Program Counter
	Memory MemoryBus // I create this interface so i can inject the memory in main.
}

// New creates a new CPU instance. The Cpu receives memory from main through dependency injection
func New(mem MemoryBus) *CPU {
	return &CPU{
		Memory: mem,
	}
}

/*

	At this point, the first thing that I thought could be useful is to start the Gameboy, so now we should
	check how does the CPU work with its registers, and also what is phisically the startup of the GBA.
	On the docs, we can actually see that we need a GBA ROM.
	9 different known official boot ROMs are known to exist, some are broken as I've seen on the docs.
	We are going to use DMG.
	After this, we are going to see how does the GBA render graphically the startup.
	I've downloaded a BootRom from this link to understand what's all about

		https://github.com/LIJI32/SameBoy/blob/master/BootROMs/dmg_boot.asm.

	Then I downloaded the bin file directly, so that i could get the bytecode version of the rom.

		curl -L -o dmg_boot.bin https://github.com/LIJI32/SameBoy/raw/master/BootROMs/prebuilt/dmg_boot.bin

*/
