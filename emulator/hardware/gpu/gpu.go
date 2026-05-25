package gpu

import (
	"fmt"
	"game/hardware/memory"
)

// GPU modes
const (
	HBlank        = 0 // Mode 0
	VBlank        = 1 // Mode 1
	OAMSearch     = 2 // Mode 2
	PixelTransfer = 3 // Mode 3
)

// GPU represents the Picture Processing Unit
type GPU struct {
	Memory      *memory.Memory // Shared memory with CPU
	Mode        uint8          // Current state machine mode
	Cycles      int            // Accumulated cycles
	CurrentLine uint8          // Current scanline (LY register)
	FrameReady  bool           // Flag for Ebitengine to draw
	ScreenData  []byte
}

// New creates a new GPU instance
func New(mem *memory.Memory) *GPU {
	return &GPU{
		Memory:     mem,
		Mode:       OAMSearch,               // Starts searching objects
		ScreenData: make([]byte, 160*144*4), // Init 160x144 RGBA buffer
	}
}

// GetTilePixel extracts the color index (0-3) for a specific x,y coordinate within a tile.
// tileID: 0-255, x: 0-7, y: 0-7
func (g *GPU) GetTilePixel(tileID uint8, x uint8, y uint8) uint8 {
	lcdc := g.Memory.Read(0xFF40)
	var tileAddress uint16

	// Bit 4: BG & Window Tile Data Select
	if (lcdc & 0x10) != 0 {
		// Modo 0x8000 (Unsigned): Tile 0-255 -> 0x8000-0x8FFF
		tileAddress = 0x8000 + (uint16(tileID) * 16)
	} else {
		// Modo 0x8800 (Signed):
		// Il Tile 0 è all'indirizzo 0x9000.
		// Gli ID 0-127 sono positivi (0x9000 in su),
		// gli ID 128-255 sono negativi (0x8800-0x8FFF).
		offset := int16(int8(tileID)) * 16
		tileAddress = uint16(int32(0x9000) + int32(offset))
	}

	rowAddress := tileAddress + (uint16(y) * 2)
	byte1 := g.Memory.Read(rowAddress)
	byte2 := g.Memory.Read(rowAddress + 1)

	lowBit := (byte1 >> (7 - x)) & 1
	highBit := (byte2 >> (7 - x)) & 1
	return (highBit << 1) | lowBit
}

// GetColorFromPalette converts the 0-3 index to RGBA based on the BGP register (0xFF47).
func (g *GPU) GetColorFromPalette(colorIndex uint8) []uint8 {
	bgp := g.Memory.Read(0xFF47)

	// Se la palette è 0, probabilmente il gioco o la Boot ROM
	// non l'hanno ancora inizializzata. Forziamo 0xE4 (standard)
	// solo per vedere cosa succede durante il caricamento.
	if bgp == 0 {
		bgp = 0xE4 // 11100100 in binario
	}

	shift := colorIndex * 2
	actualColor := (bgp >> shift) & 0x03

	switch actualColor {
	case 0:
		return []uint8{255, 255, 255, 255} // White
	case 1:
		return []uint8{192, 192, 192, 255} // Light Gray
	case 2:
		return []uint8{96, 96, 96, 255} // Dark Gray
	case 3:
		return []uint8{0, 0, 0, 255} // Black
	}

	return []uint8{0, 0, 0, 255}
}

// RenderScanline draws a single line (LY) of the background to ScreenData.

func (g *GPU) RenderScanline() {
	lcdc := g.Memory.Read(0xFF40)
	scy := g.Memory.Read(0xFF42)
	scx := g.Memory.Read(0xFF43)
	ly := g.CurrentLine

	// LOG DI CONTROLLO REGISTRI (Una volta ogni frame, alla riga 0)
	if ly == 0 {
		fmt.Printf("[GPU DEBUG] LY: %d | LCDC: %02X | SCX: %d | SCY: %d | BGP: %02X\n",
			ly, lcdc, scx, scy, g.Memory.Read(0xFF47))
	}

	// Se l'LCD è spento, non facciamo nulla
	if (lcdc & 0x80) == 0 {
		return
	}

	var tileMapBase uint16 = 0x9800
	if (lcdc & 0x08) != 0 {
		tileMapBase = 0x9C00
	}

	for screenX := uint8(0); screenX < 160; screenX++ {
		bgX := screenX + scx
		bgY := ly + scy

		tileMapCol := bgX / 8
		tileMapRow := uint16(bgY / 8)

		tileAddress := tileMapBase + (tileMapRow * 32) + uint16(tileMapCol)
		tileID := g.Memory.Read(tileAddress)

		// LOG DI CONTROLLO DATI (Se troviamo un Tile diverso da 0, lo segnaliamo)
		if tileID != 0 && ly < 144 {
			// Questo log ti dice se la CPU ha effettivamente scritto qualcosa nella mappa
			if screenX == 0 {
				fmt.Printf("[GPU DATA] Trovato TileID %d all'indirizzo 0x%04X (Riga %d)\n", tileID, tileAddress, ly)
			}
		}

		tilePixelX := bgX % 8
		tilePixelY := bgY % 8

		colorIndex := g.GetTilePixel(tileID, tilePixelX, tilePixelY)

		// LOG DI CONTROLLO PIXEL (Se il colore non è bianco, segnalalo)
		if colorIndex != 0 && ly == 70 && screenX == 80 {
			fmt.Printf("[GPU PIXEL] Colore rilevato al centro: %d\n", colorIndex)
		}

		rgba := g.GetColorFromPalette(colorIndex)

		offset := (int(ly)*160 + int(screenX)) * 4
		g.ScreenData[offset] = rgba[0]
		g.ScreenData[offset+1] = rgba[1]
		g.ScreenData[offset+2] = rgba[2]
		g.ScreenData[offset+3] = rgba[3]
	}
}

// Step advances the state machine based on CPU cycles
func (g *GPU) Step(cycles int) {
	g.Cycles += cycles

	switch g.Mode {
	case OAMSearch:
		if g.Cycles >= 80 {
			g.Cycles -= 80
			g.Mode = PixelTransfer
		}
	case PixelTransfer:
		if g.Cycles >= 172 {
			g.Cycles -= 172
			g.Mode = HBlank
			g.RenderScanline()
		}
	case HBlank:
		if g.Cycles >= 204 {
			g.Cycles -= 204
			g.CurrentLine++

			// Update LY register (0xFF44) in memory
			g.Memory.Write(0xFF44, g.CurrentLine)

			if g.CurrentLine >= 144 {
				g.Mode = VBlank
				g.FrameReady = true // Frame is ready for Ebiten
				
				// Signal V-Blank interrupt (Set bit 0 of IF register at 0xFF0F)
				if_reg := g.Memory.Read(0xFF0F)
				if_reg |= 0x01
				g.Memory.Write(0xFF0F, if_reg)
			} else {
				g.Mode = OAMSearch
			}
		}
	case VBlank:
		if g.Cycles >= 456 {
			g.Cycles -= 456
			g.CurrentLine++

			g.Memory.Write(0xFF44, g.CurrentLine)

			if g.CurrentLine > 153 {
				g.CurrentLine = 0
				g.Memory.Write(0xFF44, g.CurrentLine)
				g.Mode = OAMSearch
			}
		}
	}
}
