package main

import (
	"game/hardware/cpu"
	"game/hardware/gpu"
	"game/hardware/memory"
	"testing"
)

func TestSystemBoot(t *testing.T) {
	mem := memory.New()
	// Carica qui la tua bootrom nel buffer di mem

	ppu := gpu.New(mem)
	processor := cpu.New(mem, ppu)

	initialPC := processor.PC

	// Eseguiamo 100.000 step
	for i := 0; i < 100000; i++ {
		cycles := processor.Step()
		ppu.Step(cycles)
	}

	if processor.PC == initialPC {
		t.Fatalf("La CPU è MORTA. Il PC è ancora a 0x%04X", processor.PC)
	} else {
		t.Logf("La CPU è VIVA. PC finale: 0x%04X", processor.PC)
	}
}
