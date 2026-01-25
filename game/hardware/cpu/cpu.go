package cpu

// Here we are going to organize all the information related on the cpu
// The bible to check for documentation is https://gbdev.io/pandocs/CPU_Registers_and_Flags.html
// The gameboy cpu is composed by 8 registers at 8 bits.
//
// 	A 						-> 	Every math operation is done here
//
//  B , C , D , E , H , L 	-> 	Generic use, some instructions use these in a specific way
//
//  F 						-> 	Special register, the cpu updates this after every operation.

// We can think about the cpu through a struct:

type CPU struct {
	// 8 bit registers
	A, F uint8
	B, C uint8
	D, E uint8
	H, L uint8

	// 16-bit registers
	SP uint16 // Stack Pointer
	PC uint16 // Program Counter

	// Memory pointer should exist in future
	// memory *Memory

	// Clock cycles
	cycles uint64
}
