package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"game/hardware/cpu"
	"game/hardware/gpu" // Added GPU package
	"game/hardware/memory"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const DMG_BOOT_PATH = "/Users/lorenzomazzoccante/Desktop/gba/emulator/hardware/assets/dmg_boot.bin"

type Game struct {
	selectedRow int
	gbaFiles    GBAFiles
	cpu         *cpu.CPU
	ppu         *gpu.GPU // Added PPU reference
	mem         *memory.Memory
	romLoaded   bool
	gbScreen    *ebiten.Image // Image for the Game Boy screen
}

type GBAFiles struct {
	Files      []string
	Name       string
	isSelected bool
}

// Find files once
func FindGBAFiles(dir string) (GBAFiles, error) {
	var result GBAFiles

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(d.Name()) == ".gb" {
			result.Files = append(result.Files, path)
			result.Name = filepath.Base(path)
		}

		return nil
	})

	return result, err
}

func hexDumpToFile(data []byte, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	for i := 0; i < len(data); i += 16 {
		fmt.Fprintf(f, "%08X  ", i)
		for j := 0; j < 16 && i+j < len(data); j++ {
			fmt.Fprintf(f, "%02X ", data[i+j])
		}
		fmt.Fprint(f, " ")
		for j := 0; j < 16 && i+j < len(data); j++ {
			b := data[i+j]
			if b >= 32 && b <= 126 {
				fmt.Fprintf(f, "%c", b)
			} else {
				fmt.Fprint(f, ".")
			}
		}
		fmt.Fprintln(f)
	}
	return nil
}

// Update executes UI logic or emulator loop
func (g *Game) Update() error {
	// 1. UI Mode: Select and load ROM
	if !g.romLoaded {
		if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
			if g.selectedRow < len(g.gbaFiles.Files)-1 {
				g.selectedRow++
			}
		}

		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			if g.selectedRow > 0 {
				g.selectedRow--
			}
		}

		if ebiten.IsKeyPressed(ebiten.KeyEnter) && len(g.gbaFiles.Files) > 0 {
			selectedFile := g.gbaFiles.Files[g.selectedRow]
			data, err := os.ReadFile(selectedFile)
			if err != nil {
				log.Fatal(err)
			}

			// 1. Carica la ROM correttamente
			g.mem.LoadROM(data)

			// 2. INIZIALIZZAZIONE REGISTRI (Stato post-boot standard)
			g.cpu.A = 0x01
			g.cpu.F = 0xB0 // Z=1, H=1, C=1
			g.cpu.B = 0x00
			g.cpu.C = 0x13
			g.cpu.D = 0x00
			g.cpu.E = 0xD8
			g.cpu.H = 0x01
			g.cpu.L = 0x4D
			g.cpu.SP = 0xFFFE
			g.cpu.PC = 0x0100 // SALTA AL GIOCO

			// 3. REGISTRI HARDWARE (Stato post-boot standard)
			g.mem.Write(0xFF10, 0x80) // Sound
			g.mem.Write(0xFF11, 0xBF)
			g.mem.Write(0xFF12, 0xF3)
			g.mem.Write(0xFF14, 0xBF)
			g.mem.Write(0xFF16, 0x3F)
			g.mem.Write(0xFF19, 0xBF)
			g.mem.Write(0xFF1A, 0x7F)
			g.mem.Write(0xFF1B, 0xFF)
			g.mem.Write(0xFF1C, 0x9F)
			g.mem.Write(0xFF1E, 0xBF)
			g.mem.Write(0xFF20, 0xFF)
			g.mem.Write(0xFF23, 0xBF)
			g.mem.Write(0xFF24, 0x77)
			g.mem.Write(0xFF25, 0xF3)
			g.mem.Write(0xFF26, 0xF1)
			g.mem.Write(0xFF40, 0x91) // LCD ON
			g.mem.Write(0xFF47, 0xE4) // Palette

			g.romLoaded = true
			return nil
		}
		return nil
	}

	// 2. Emulator Mode: Sync CPU and PPU
	for !g.ppu.FrameReady {
		cycles := g.cpu.Step()
		g.ppu.Step(cycles)
	}

	g.ppu.FrameReady = false
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.romLoaded {
		// 1. Render Game Boy Screen
		if g.gbScreen == nil {
			g.gbScreen = ebiten.NewImage(160, 144)
		}
		g.gbScreen.WritePixels(g.ppu.ScreenData)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(10, 10)
		screen.DrawImage(g.gbScreen, op)

		// 2. Render Debugger Panel
		x := 185
		ebitenutil.DebugPrintAt(screen, "--- REGISTERS ---", x, 10)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("A: %02X  F: %02X", g.cpu.A, g.cpu.F), x, 25)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("B: %02X  C: %02X", g.cpu.B, g.cpu.C), x, 40)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("D: %02X  E: %02X", g.cpu.D, g.cpu.E), x, 55)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("H: %02X  L: %02X", g.cpu.H, g.cpu.L), x, 70)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SP: %04X", g.cpu.SP), x, 90)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("PC: %04X", g.cpu.PC), x, 105)

		ebitenutil.DebugPrintAt(screen, "--- FLAGS ---", x, 130)
		z := (g.cpu.F >> 7) & 1
		n := (g.cpu.F >> 6) & 1
		h := (g.cpu.F >> 5) & 1
		c := (g.cpu.F >> 4) & 1
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Z:%d N:%d H:%d C:%d", z, n, h, c), x, 145)

		ebitenutil.DebugPrintAt(screen, "--- SYSTEM ---", x, 170)
		ime := 0
		if g.cpu.IME {
			ime = 1
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("IME: %d", ime), x, 185)
		return
	}

	for i, file := range g.gbaFiles.Files {
		prefix := "  "
		if i == g.selectedRow {
			prefix = "> "
		}
		ebitenutil.DebugPrintAt(screen, prefix+file, 10, 20+i*20)
	}

	if len(g.gbaFiles.Files) == 0 {
		ebitenutil.DebugPrint(screen, "No .gb files found in ./roms")
	} else {
		ebitenutil.DebugPrintAt(screen, "Use arrows and Enter to select", 10, 10)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Sophiagb")

	// 1. Init hardware (Una sola istanza per tipo)
	mem := memory.New()
	ppu := gpu.New(mem)
	cpuInstance := cpu.New(mem, ppu)

	// 2. Caricamento Boot ROM
	err := mem.LoadBootROM(DMG_BOOT_PATH)
	if err != nil {
		fmt.Println("Warning: Boot ROM not loaded:", err)
	}

	// 3. Load file list
	files, _ := FindGBAFiles("./roms")

	game := &Game{
		gbaFiles: files,
		cpu:      cpuInstance,
		ppu:      ppu, // USA LA STESSA ISTANZA DELLA CPU
		mem:      mem,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
