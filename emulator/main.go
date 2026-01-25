package main

// I've created this file based on the official docs of ebitengine
//
// https://ebitengine.org/en/documents/install.html
//
// This is used to create a window that displays the emulator

// https://gbdev.io/pandocs/About.html <--- Emulator Bible to consult :)
import (
	"fmt"

	"game/hardware/cpu"
	"game/hardware/memory"

	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	selectedRow int
}

type GBAFiles struct {
	Files      []string
	Name       string
	isSelected bool
}

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

// Debug functions to understand what a rom is all about in terms of bytecode.
func hexDump(data []byte, start, length int) {
	end := start + length
	if end > len(data) {
		end = len(data)
	}

	for i := start; i < end; i += 16 {
		fmt.Printf("%08X  ", i)

		for j := 0; j < 16 && i+j < end; j++ {
			fmt.Printf("%02X ", data[i+j])
		}

		fmt.Printf(" ")

		for j := 0; j < 16 && i+j < end; j++ {
			b := data[i+j]
			if b >= 32 && b <= 126 {
				fmt.Printf("%c", b)
			} else {
				fmt.Print(".")
			}
		}

		fmt.Println()
	}
}

func (g *Game) Update() error {

	gbaFiles, _ := FindGBAFiles("./roms")

	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		if g.selectedRow < len(gbaFiles.Files)-1 {
			g.selectedRow++
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		if g.selectedRow > 0 {
			g.selectedRow--
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyEnter) && len(gbaFiles.Files) > 0 {
		selectedFile := gbaFiles.Files[g.selectedRow]
		// Here you have selected your game,so we must open it
		open, err := ebitenutil.OpenFile(selectedFile)
		data, err := io.ReadAll(open)
		if err != nil {
			log.Fatal(err)
		}

		//fmt.Println(data)
		fmt.Printf("Game: %s\n ", data[0xA0:0xAC]) //It is a standard for gba roms to print the game's name in these bytes.
		hexDumpToFile(data, "pokemon_dump.txt")    // I've used this to get a proper rapresentation in bytecode of what is a rom all about.
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	gbaFiles, _ := FindGBAFiles("./roms")

	for i, file := range gbaFiles.Files {
		prefix := "  "
		if i == g.selectedRow {
			prefix = "> " // This is the arrow that shows which game you have selected
		}
		ebitenutil.DebugPrintAt(screen, prefix+file, 10, 20+i*20)
	}

	if len(gbaFiles.Files) == 0 {
		ebitenutil.DebugPrint(screen, "Nessun file .gb trovato in ./roms")
	} else {
		ebitenutil.DebugPrintAt(screen, "Usa freccia in basso e Invio per selezionare", 10, 10)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Sophiagb")

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
	// Dependency injection to use cpu and memory.
	// We initialize these 2 things so we can even test.

	mem := memory.New()
	cpu := cpu.New(mem)

}
