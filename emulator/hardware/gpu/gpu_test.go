package gpu

import (
	"game/hardware/memory"
	"testing"
)

func TestGPURenderLogo(t *testing.T) {
	// 1. Setup Memoria e GPU
	mem := memory.New()
	ppu := New(mem)

	// 2. Mock dei dati: Carichiamo un Tile "a scacchiera" all'indirizzo 0x8000
	// Ogni Tile è 16 byte. 0x55 e 0xAA creano pattern alternati.
	tileData := []byte{
		0x55, 0x00, 0x55, 0x00, 0x55, 0x00, 0x55, 0x00,
		0x55, 0x00, 0x55, 0x00, 0x55, 0x00, 0x55, 0x00,
	}
	for i, b := range tileData {
		mem.Write(uint16(0x8000+i), b)
	}

	// 3. Mock della Tile Map: Riempiamo la prima riga della mappa con il Tile 0
	// La mappa inizia a 0x9800. Mettiamo il Tile 0 nelle prime 32 posizioni.
	for i := uint16(0); i < 32; i++ {
		mem.Write(0x9800+i, 0)
	}

	// 4. Configurazione Registri Hardware
	mem.Write(0xFF40, 0x91) // LCD On, BG On, Tile Data 0x8000
	mem.Write(0xFF47, 0xE4) // Palette standard (11100100)
	mem.Write(0xFF42, 0)    // Scroll Y = 0
	mem.Write(0xFF43, 0)    // Scroll X = 0

	// 5. Eseguiamo il rendering della prima riga (LY = 0)
	ppu.CurrentLine = 0
	ppu.RenderScanline()

	// 6. Verifica: Il primo pixel (0,0) non deve essere vuoto (0,0,0,0)
	// Deve essere uno dei colori della palette (es. Bianco 255,255,255,255)
	pixelOffset := 0
	r := ppu.ScreenData[pixelOffset]
	g := ppu.ScreenData[pixelOffset+1]
	b := ppu.ScreenData[pixelOffset+2]
	a := ppu.ScreenData[pixelOffset+3]

	if a == 0 {
		t.Errorf("Il buffer è ancora vuoto! Alpha è 0 al pixel 0,0")
	}

	t.Logf("Pixel 0,0: R:%d G:%d B:%d A:%d", r, g, b, a)
}
