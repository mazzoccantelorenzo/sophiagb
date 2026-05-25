package cpu

import (
	"game/hardware/gpu"
	"game/hardware/memory"
	"os"
	"testing"
)

func TestFullSystemStartup(t *testing.T) {
	// 1. Inizializza i componenti reali
	mem := memory.New()
	ppu := gpu.New(mem)

	// Passiamo mem e ppu alla CPU (Dependency Injection)
	processor := New(mem, ppu)

	// 2. Apri il vero file della Boot ROM o di una ROM
	// Assicurati che il percorso sia corretto rispetto a dove lanci il test
	romData, err := os.ReadFile("../../roms/Pokemon.gb")
	if err != nil {
		t.Fatalf("Impossibile aprire il file ROM: %v", err)
	}

	// 3. Carica la ROM nella memoria (simulando il caricamento hardware)
	for i, b := range romData {
		mem.Write(uint16(i), b)
	}

	// 4. Esegui il sistema per un numero X di istruzioni
	// Vogliamo vedere se il PC (Program Counter) avanza e se la GPU accumula cicli
	initialPC := processor.PC
	iterations := 5000

	for i := 0; i < iterations; i++ {
		// Recuperiamo l'opcode per il log in caso di errore
		currentPC := processor.PC

		// Eseguiamo uno step
		cycles := processor.Step()

		// Verifica che il PC si sia mosso (evitiamo loop infiniti su opcode non implementati)
		if processor.PC == currentPC && cycles == 0 {
			t.Fatalf("CPU bloccata al PC 0x%04X (Opcode sconosciuto?)", currentPC)
		}
	}

	// 5. Verifiche finali
	t.Logf("Test completato dopo %d iterazioni.", iterations)
	t.Logf("PC finale: 0x%04X", processor.PC)
	t.Logf("Cicli totali accumulati nella GPU: %d", ppu.Cycles)

	if processor.PC == initialPC {
		t.Error("Errore: Il Program Counter non è mai avanzato!")
	}

	// Controllo critico: la CPU ha scritto nel registro LCDC (0xFF40)?
	lcdc := mem.Read(0xFF40)
	t.Logf("Valore finale registro LCDC: 0x%02X", lcdc)
}
